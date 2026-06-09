import { useEffect, useMemo, useState } from 'react';
import { AlertCircle, BookOpen, FileText, Loader2 } from 'lucide-react';
import {
  EpisodeDetail,
  PodcastListItem,
  SummarySegment,
  TranscriptSegment,
  fetchEpisodeDetail
} from '../services/podcastClient';

type Props = {
  selectedPodcast?: PodcastListItem;
};

type RequestState = 'idle' | 'loading' | 'ready' | 'error';

type TranscriptSummaryGroup = {
  key: string;
  orderIndex: number;
  summary?: SummarySegment;
  transcripts: TranscriptSegment[];
};

function statusLabel(status?: string) {
  if (!status) return 'Not started';
  return status.replace(/_/g, ' ');
}

function isFailed(detail?: EpisodeDetail) {
  const status = (detail?.latestJob?.status || detail?.status || '').toLowerCase();
  return status === 'failed';
}

function formatTimeRange(segment: TranscriptSegment) {
  if (segment.startSeconds === undefined && segment.endSeconds === undefined) return '';
  const format = (seconds?: number) => {
    if (seconds === undefined) return '';
    const minutes = Math.floor(seconds / 60);
    const remainder = Math.floor(seconds % 60)
      .toString()
      .padStart(2, '0');
    return `${minutes}:${remainder}`;
  };
  const start = format(segment.startSeconds);
  const end = format(segment.endSeconds);
  return start && end ? `${start}-${end}` : start || end;
}

function groupTranscriptBySummary(detail?: EpisodeDetail): TranscriptSummaryGroup[] {
  if (!detail) return [];
  const transcripts = [...detail.transcriptSegments].sort((a, b) => a.orderIndex - b.orderIndex);
  const transcriptByID = new Map(transcripts.map((item) => [item.id, item]));
  const usedTranscriptIDs = new Set<string>();

  const summaryGroups = [...detail.summarySegments]
    .sort((a, b) => a.orderIndex - b.orderIndex)
    .map((summary) => {
      const mappedTranscripts = summary.sourceTranscriptSegmentIds
        .map((id) => transcriptByID.get(id))
        .filter((item): item is TranscriptSegment => Boolean(item))
        .sort((a, b) => a.orderIndex - b.orderIndex);
      mappedTranscripts.forEach((item) => usedTranscriptIDs.add(item.id));
      return {
        key: summary.id || `summary-${summary.orderIndex}`,
        orderIndex: mappedTranscripts[0]?.orderIndex ?? summary.orderIndex,
        summary,
        transcripts: mappedTranscripts
      };
    });

  const orphanGroups = transcripts
    .filter((item) => !usedTranscriptIDs.has(item.id))
    .map((item) => ({
      key: item.id || `transcript-${item.orderIndex}`,
      orderIndex: item.orderIndex,
      transcripts: [item]
    }));

  return [...summaryGroups, ...orphanGroups].sort((a, b) => a.orderIndex - b.orderIndex);
}

export function PodcastView({ selectedPodcast }: Props) {
  const [requestState, setRequestState] = useState<RequestState>('idle');
  const [detail, setDetail] = useState<EpisodeDetail | undefined>();
  const [error, setError] = useState<string | undefined>();

  useEffect(() => {
    if (!selectedPodcast?.id) {
      setRequestState('idle');
      setDetail(undefined);
      setError(undefined);
      return;
    }

    let cancelled = false;
    setRequestState('loading');
    setError(undefined);
    fetchEpisodeDetail(selectedPodcast.id)
      .then((nextDetail) => {
        if (cancelled) return;
        setDetail(nextDetail);
        setRequestState('ready');
      })
      .catch((e) => {
        if (cancelled) return;
        setDetail(undefined);
        setError(e instanceof Error ? e.message.trim() : 'Unable to load episode detail');
        setRequestState('error');
      });

    return () => {
      cancelled = true;
    };
  }, [selectedPodcast?.id]);

  const groups = useMemo(() => groupTranscriptBySummary(detail), [detail]);
  const failed = isFailed(detail);
  const latestError = detail?.latestJob?.errorMessage;
  const hasTranscript = groups.some((group) => group.transcripts.length > 0);
  const title = detail?.title || selectedPodcast?.title || selectedPodcast?.url || 'Episode viewer';

  if (!selectedPodcast?.id) {
    return (
      <section className="rounded border bg-white p-5 shadow-sm">
        <div className="flex items-start gap-3 text-slate-700">
          <BookOpen aria-hidden="true" className="mt-0.5 h-5 w-5" />
          <div>
            <h2 className="text-xl font-semibold text-slate-950">Demo transcript and summary</h2>
            <p className="mt-1 text-sm">Select an episode to view transcript segments and mapped summaries.</p>
          </div>
        </div>
      </section>
    );
  }

  return (
    <section className="rounded border bg-white p-5 shadow-sm">
      <div className="mb-5 flex items-start justify-between gap-4">
        <div className="min-w-0">
          <h2 className="text-xl font-semibold text-slate-950">Demo transcript and summary</h2>
          <p className="mt-1 truncate text-sm text-slate-600" title={title}>
            {title}
          </p>
        </div>
        {detail && (
          <span
            className={`shrink-0 rounded-full px-3 py-1 text-xs font-semibold capitalize ${
              failed
                ? 'bg-red-50 text-red-700'
                : hasTranscript
                ? 'bg-emerald-50 text-emerald-700'
                : 'bg-amber-50 text-amber-700'
            }`}
          >
            {statusLabel(detail.latestJob?.status || detail.status)}
          </span>
        )}
      </div>

      {requestState === 'loading' && (
        <div className="flex items-center gap-2 rounded border border-blue-200 bg-blue-50 px-3 py-2 text-sm text-blue-800">
          <Loader2 aria-hidden="true" className="h-4 w-4 animate-spin" />
          Loading episode detail...
        </div>
      )}

      {requestState === 'error' && (
        <div role="alert" className="flex items-start gap-2 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          <AlertCircle aria-hidden="true" className="mt-0.5 h-4 w-4" />
          {error || 'Unable to load episode detail'}
        </div>
      )}

      {requestState === 'ready' && failed && (
        <div role="alert" className="mb-4 rounded border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          <p className="font-medium">Processing failed</p>
          {latestError && <p className="mt-1">{latestError}</p>}
        </div>
      )}

      {requestState === 'ready' && !hasTranscript && (
        <div className="rounded border border-slate-200 bg-slate-50 px-3 py-4 text-sm text-slate-700">
          No transcript segments available yet.
        </div>
      )}

      {requestState === 'ready' && hasTranscript && (
        <div className="space-y-4">
          {groups.map((group, index) => (
            <article key={group.key} className="grid gap-3 rounded border border-slate-200 p-4 md:grid-cols-[minmax(0,1.2fr)_minmax(220px,0.8fr)]">
              <div>
                <div className="mb-2 flex items-center gap-2 text-xs font-semibold uppercase text-slate-500">
                  <FileText aria-hidden="true" className="h-4 w-4" />
                  Transcript group {index + 1}
                </div>
                <div className="space-y-3">
                  {group.transcripts.map((segment) => (
                    <div key={segment.id} className="rounded border border-slate-100 bg-slate-50 px-3 py-2">
                      <div className="mb-1 flex items-center justify-between gap-2 text-[11px] font-medium text-slate-500">
                        <span>Segment {segment.orderIndex}</span>
                        {formatTimeRange(segment) && <span>{formatTimeRange(segment)}</span>}
                      </div>
                      <p className="whitespace-pre-line break-words text-sm leading-6 text-slate-900">{segment.text}</p>
                    </div>
                  ))}
                </div>
              </div>
              <aside className="rounded border border-indigo-100 bg-indigo-50 px-3 py-3 text-sm text-indigo-950">
                <div className="mb-2 text-xs font-semibold uppercase text-indigo-700">Summary</div>
                {group.summary ? (
                  <p className="whitespace-pre-line break-words leading-6">{group.summary.text}</p>
                ) : (
                  <p className="text-indigo-800">No summary is mapped to this transcript segment yet.</p>
                )}
              </aside>
            </article>
          ))}
        </div>
      )}
    </section>
  );
}
