import { apiRequest } from './client';

export type JobAccepted = {
  jobId?: string;
  podcastId: string;
  status: string;
  existing: boolean;
  latestStatus?: string;
};

export type EpisodeJobStatus = {
  jobId: string;
  podcastId: string;
  type: string;
  status: string;
  startedAt?: string;
  completedAt?: string;
  durationMs?: number;
  errorMessage?: string | null;
};

export type EpisodeStatus = {
  podcastId: string;
  id: string;
  url: string;
  title: string;
  hasTranscript: boolean;
  status: string;
  latestJob?: EpisodeJobStatus;
};

type Paragraph = {
  paragraphId: string;
  orderIndex: number;
  text: string;
  summary: string;
};

export type TranscriptSegment = {
  id: string;
  orderIndex: number;
  text: string;
  startSeconds?: number;
  endSeconds?: number;
};

export type SummarySegment = {
  id: string;
  orderIndex: number;
  text: string;
  sourceTranscriptSegmentIds: string[];
};

export type EpisodeDetail = EpisodeStatus & {
  description?: string;
  transcriptSegments: TranscriptSegment[];
  summarySegments: SummarySegment[];
  paragraphs: Paragraph[];
};

type TranscriptView = {
  podcastId: string;
  paragraphs: Paragraph[];
};

export type PodcastListItem = {
  id: string;
  url: string;
  title: string;
  hasTranscript: boolean;
  createdAt: string;
  latestJobId?: string;
  latestStatus?: string;
};

export async function ingestPodcast(url: string): Promise<JobAccepted> {
  const resp = await apiRequest<unknown>('/api/podcasts/ingest', {
    method: 'POST',
    body: { url }
  });
  return normalizeJobAccepted(resp);
}

export async function fetchView(podcastId: string): Promise<TranscriptView> {
  return apiRequest<TranscriptView>(`/api/podcasts/${podcastId}/view`);
}

export async function fetchEpisodeDetail(podcastId: string): Promise<EpisodeDetail> {
  const resp = await apiRequest<unknown>(`/api/podcasts/${podcastId}`);
  return normalizeEpisodeDetail(resp);
}

export async function fetchEpisodeStatus(podcastId: string): Promise<EpisodeStatus> {
  const resp = await apiRequest<unknown>(`/api/podcasts/${podcastId}/status`);
  return normalizeEpisodeStatus(resp);
}

const asRecord = (raw: unknown): Record<string, unknown> =>
  raw && typeof raw === 'object' ? (raw as Record<string, unknown>) : {};

const pickString = (record: Record<string, unknown>, ...keys: string[]) => {
  for (const key of keys) {
    const v = record[key];
    if (typeof v === 'string') return v;
  }
  return '';
};

const pickBool = (record: Record<string, unknown>, ...keys: string[]) => {
  for (const key of keys) {
    const v = record[key];
    if (typeof v === 'boolean') return v;
  }
  return false;
};

const pickNumber = (record: Record<string, unknown>, ...keys: string[]) => {
  for (const key of keys) {
    const v = record[key];
    if (typeof v === 'number') return v;
  }
  return undefined;
};

function normalizeJobAccepted(raw: unknown): JobAccepted {
  const record = asRecord(raw);
  const latestStatus = pickString(record, 'latestStatus', 'LatestStatus');
  const jobId = pickString(record, 'jobId', 'JobID');
  return {
    jobId: jobId || undefined,
    podcastId: pickString(record, 'podcastId', 'PodcastID'),
    status: pickString(record, 'status', 'Status'),
    existing: pickBool(record, 'existing', 'Existing'),
    latestStatus: latestStatus || undefined
  };
}

function normalizeJobStatus(raw: unknown): EpisodeJobStatus | undefined {
  const record = asRecord(raw);
  const jobId = pickString(record, 'jobId', 'JobID');
  if (!jobId) return undefined;
  const errorValue = record.errorMessage ?? record.ErrorMessage;
  return {
    jobId,
    podcastId: pickString(record, 'podcastId', 'PodcastID'),
    type: pickString(record, 'type', 'Type'),
    status: pickString(record, 'status', 'Status'),
    startedAt: pickString(record, 'startedAt', 'StartedAt') || undefined,
    completedAt: pickString(record, 'completedAt', 'CompletedAt') || undefined,
    durationMs: pickNumber(record, 'durationMs', 'DurationMs'),
    errorMessage: typeof errorValue === 'string' ? errorValue : null
  };
}

function normalizeEpisodeStatus(raw: unknown): EpisodeStatus {
  const record = asRecord(raw);
  const latestJob = normalizeJobStatus(record.latestJob ?? record.LatestJob);
  return {
    podcastId: pickString(record, 'podcastId', 'PodcastID'),
    id: pickString(record, 'id', 'ID') || pickString(record, 'podcastId', 'PodcastID'),
    url: pickString(record, 'url', 'URL'),
    title: pickString(record, 'title', 'Title') || pickString(record, 'url', 'URL'),
    hasTranscript: pickBool(record, 'hasTranscript', 'HasTranscript'),
    status: pickString(record, 'status', 'Status') || latestJob?.status || 'not_started',
    latestJob
  };
}

function normalizeTranscriptSegment(raw: unknown): TranscriptSegment {
  const record = asRecord(raw);
  return {
    id: pickString(record, 'id', 'ID'),
    orderIndex: pickNumber(record, 'orderIndex', 'OrderIndex') ?? 0,
    text: pickString(record, 'text', 'Text'),
    startSeconds: pickNumber(record, 'startSeconds', 'StartSeconds'),
    endSeconds: pickNumber(record, 'endSeconds', 'EndSeconds')
  };
}

function normalizeSummarySegment(raw: unknown): SummarySegment {
  const record = asRecord(raw);
  const rawSourceIds = record.sourceTranscriptSegmentIds ?? record.SourceTranscriptSegmentIDs;
  return {
    id: pickString(record, 'id', 'ID'),
    orderIndex: pickNumber(record, 'orderIndex', 'OrderIndex') ?? 0,
    text: pickString(record, 'text', 'Text'),
    sourceTranscriptSegmentIds: Array.isArray(rawSourceIds)
      ? rawSourceIds.filter((item): item is string => typeof item === 'string')
      : []
  };
}

function normalizeParagraph(raw: unknown): Paragraph {
  const record = asRecord(raw);
  return {
    paragraphId: pickString(record, 'paragraphId', 'ParagraphID'),
    orderIndex: pickNumber(record, 'orderIndex', 'OrderIndex') ?? 0,
    text: pickString(record, 'text', 'Text'),
    summary: pickString(record, 'summary', 'Summary')
  };
}

function normalizeEpisodeDetail(raw: unknown): EpisodeDetail {
  const record = asRecord(raw);
  const status = normalizeEpisodeStatus(raw);
  const transcriptSegments = record.transcriptSegments ?? record.TranscriptSegments;
  const summarySegments = record.summarySegments ?? record.SummarySegments;
  const paragraphs = record.paragraphs ?? record.Paragraphs;
  return {
    ...status,
    description: pickString(record, 'description', 'Description') || undefined,
    transcriptSegments: Array.isArray(transcriptSegments) ? transcriptSegments.map(normalizeTranscriptSegment) : [],
    summarySegments: Array.isArray(summarySegments) ? summarySegments.map(normalizeSummarySegment) : [],
    paragraphs: Array.isArray(paragraphs) ? paragraphs.map(normalizeParagraph) : []
  };
}

function normalizeItem(raw: unknown): PodcastListItem {
  const record = asRecord(raw);
  const latestJobId = pickString(record, 'latestJobId', 'LatestJobID');
  const latestStatus = pickString(record, 'latestStatus', 'LatestStatus');
  return {
    id: pickString(record, 'id', 'ID'),
    url: pickString(record, 'url', 'URL'),
    title: pickString(record, 'title', 'Title') || pickString(record, 'url', 'URL'),
    hasTranscript: pickBool(record, 'hasTranscript', 'HasTranscript'),
    createdAt: pickString(record, 'createdAt', 'CreatedAt'),
    latestJobId: latestJobId || undefined,
    latestStatus: latestStatus || undefined
  };
}

export async function fetchPodcasts(): Promise<PodcastListItem[]> {
  const resp = await apiRequest<{ items: unknown[] }>('/api/podcasts');
  return Array.isArray(resp.items) ? resp.items.map(normalizeItem) : [];
}
