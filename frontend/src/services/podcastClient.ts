import { apiRequest } from './client';

type JobAccepted = {
  jobId: string;
  podcastId: string;
  status: string;
};

type Paragraph = {
  paragraphId: string;
  orderIndex: number;
  text: string;
  summary: string;
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
  return apiRequest<JobAccepted>('/api/podcasts/ingest', {
    method: 'POST',
    body: { url }
  });
}

export async function fetchView(podcastId: string): Promise<TranscriptView> {
  return apiRequest<TranscriptView>(`/api/podcasts/${podcastId}/view`);
}

function normalizeItem(raw: unknown): PodcastListItem {
  const record: Record<string, unknown> = raw && typeof raw === 'object' ? (raw as Record<string, unknown>) : {};
  const pickString = (...keys: string[]) => {
    for (const key of keys) {
      const v = record[key];
      if (typeof v === 'string') return v;
    }
    return '';
  };
  const pickBool = (...keys: string[]) => {
    for (const key of keys) {
      const v = record[key];
      if (typeof v === 'boolean') return v;
    }
    return false;
  };
  const latestJobId = pickString('latestJobId', 'LatestJobID');
  const latestStatus = pickString('latestStatus', 'LatestStatus');
  return {
    id: pickString('id', 'ID'),
    url: pickString('url', 'URL'),
    title: pickString('title', 'Title') || pickString('url', 'URL'),
    hasTranscript: pickBool('hasTranscript', 'HasTranscript'),
    createdAt: pickString('createdAt', 'CreatedAt'),
    latestJobId: latestJobId || undefined,
    latestStatus: latestStatus || undefined
  };
}

export async function fetchPodcasts(): Promise<PodcastListItem[]> {
  const resp = await apiRequest<{ items: unknown[] }>('/api/podcasts');
  return Array.isArray(resp.items) ? resp.items.map(normalizeItem) : [];
}
