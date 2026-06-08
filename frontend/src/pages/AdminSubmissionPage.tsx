import { useEffect, useMemo, useState } from 'react';
import { AlertCircle, CheckCircle2, RefreshCw, Send, Timer } from 'lucide-react';
import { UrlInput } from '../components/UrlInput';
import {
  EpisodeStatus,
  JobAccepted,
  PodcastListItem,
  fetchEpisodeStatus,
  ingestPodcast
} from '../services/podcastClient';

type Props = {
  selectedPodcast?: PodcastListItem;
  onSubmissionRecorded?: () => void;
};

type RequestState = 'idle' | 'submitting' | 'loading-status' | 'ready' | 'error';
type SubmissionKind = 'new' | 'existing' | 'selected' | undefined;

function statusLabel(status?: string) {
  if (!status) return 'Not started';
  return status.replace(/_/g, ' ');
}

function statusTone(status?: string) {
  const normalized = (status || '').toLowerCase();
  if (normalized === 'succeeded' || normalized === 'completed') return 'border-emerald-200 bg-emerald-50 text-emerald-800';
  if (normalized === 'failed') return 'border-red-200 bg-red-50 text-red-800';
  if (normalized === 'queued' || normalized === 'running' || normalized === 'downloading' || normalized === 'chunking' || normalized === 'transcribing' || normalized === 'summarizing') {
    return 'border-blue-200 bg-blue-50 text-blue-800';
  }
  return 'border-slate-200 bg-slate-50 text-slate-800';
}

function statusIcon(status?: string) {
  const normalized = (status || '').toLowerCase();
  if (normalized === 'succeeded' || normalized === 'completed') return <CheckCircle2 aria-hidden="true" className="h-5 w-5" />;
  if (normalized === 'failed') return <AlertCircle aria-hidden="true" className="h-5 w-5" />;
  return <Timer aria-hidden="true" className="h-5 w-5" />;
}

function seededStatusFromIngest(result: JobAccepted, submittedURL: string): EpisodeStatus {
  return {
    podcastId: result.podcastId,
    id: result.podcastId,
    url: submittedURL,
    title: submittedURL,
    hasTranscript: false,
    status: result.latestStatus || result.status || 'queued',
    latestJob: result.jobId
      ? {
          jobId: result.jobId,
          podcastId: result.podcastId,
          type: 'ingest',
          status: result.latestStatus || result.status || 'queued'
        }
      : undefined
  };
}

export function AdminSubmissionPage({ selectedPodcast, onSubmissionRecorded }: Props) {
  const [requestState, setRequestState] = useState<RequestState>('idle');
  const [submissionKind, setSubmissionKind] = useState<SubmissionKind>();
  const [episodeStatus, setEpisodeStatus] = useState<EpisodeStatus | undefined>();
  const [error, setError] = useState<string | undefined>();

  const loadStatus = async (podcastId: string, kind: SubmissionKind) => {
    setRequestState('loading-status');
    setError(undefined);
    setSubmissionKind(kind);
    try {
      const status = await fetchEpisodeStatus(podcastId);
      setEpisodeStatus(status);
      setRequestState('ready');
    } catch (e) {
      setRequestState('error');
      setError(e instanceof Error ? e.message.trim() : 'Unable to load episode status');
    }
  };

  useEffect(() => {
    if (selectedPodcast?.id) {
      void loadStatus(selectedPodcast.id, 'selected');
    }
  }, [selectedPodcast?.id]);

  const handleSubmit = async (url: string) => {
    setRequestState('submitting');
    setSubmissionKind(undefined);
    setError(undefined);
    setEpisodeStatus(undefined);
    try {
      const result = await ingestPodcast(url);
      setSubmissionKind(result.existing ? 'existing' : 'new');
      setEpisodeStatus(seededStatusFromIngest(result, url));
      onSubmissionRecorded?.();
      await loadStatus(result.podcastId, result.existing ? 'existing' : 'new');
    } catch (e) {
      setRequestState('error');
      setError(e instanceof Error ? e.message.trim() : 'Unable to submit podcast URL');
    }
  };

  const headline = useMemo(() => {
    if (submissionKind === 'existing') return 'Existing episode found';
    if (submissionKind === 'new') return 'New processing job started';
    if (submissionKind === 'selected') return 'Episode status';
    return 'Submit a podcast URL';
  }, [submissionKind]);

  const latestJob = episodeStatus?.latestJob;
  const displayStatus = latestJob?.status || episodeStatus?.status;
  const isBusy = requestState === 'submitting' || requestState === 'loading-status';

  return (
    <section className="space-y-5">
      <div className="rounded border bg-white p-5 shadow-sm">
        <div className="mb-4 flex items-start justify-between gap-4">
          <div>
            <h2 className="text-xl font-semibold text-slate-950">Admin submission</h2>
            <p className="mt-1 text-sm text-slate-600">Submit a podcast URL and inspect its local processing state.</p>
          </div>
          {episodeStatus?.podcastId && (
            <button
              type="button"
              onClick={() => void loadStatus(episodeStatus.podcastId, submissionKind)}
              className="inline-flex h-9 items-center gap-2 rounded border border-slate-300 px-3 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-50"
              disabled={isBusy}
            >
              <RefreshCw aria-hidden="true" className="h-4 w-4" />
              Refresh
            </button>
          )}
        </div>

        <UrlInput onSubmit={handleSubmit} status={isBusy ? 'loading' : requestState === 'error' ? 'error' : 'idle'} errorMessage={error} />
      </div>

      <div className={`rounded border p-5 shadow-sm ${episodeStatus ? statusTone(displayStatus) : 'border-slate-200 bg-white text-slate-800'}`}>
        <div className="flex items-start gap-3">
          <div className="mt-0.5">{episodeStatus ? statusIcon(displayStatus) : <Send aria-hidden="true" className="h-5 w-5" />}</div>
          <div className="min-w-0 flex-1 space-y-3">
            <div>
              <h3 className="text-lg font-semibold">{headline}</h3>
              {submissionKind === 'existing' && (
                <p className="mt-1 text-sm">This URL is already in the local library; no new processing run was implied.</p>
              )}
              {submissionKind === 'new' && (
                <p className="mt-1 text-sm">The backend accepted the URL and created a processing job.</p>
              )}
              {!episodeStatus && requestState === 'idle' && (
                <p className="mt-1 text-sm text-slate-600">No episode selected yet.</p>
              )}
            </div>

            {episodeStatus && (
              <dl className="grid gap-3 text-sm sm:grid-cols-2">
                <div>
                  <dt className="font-medium">Processing stage</dt>
                  <dd className="capitalize">{statusLabel(displayStatus)}</dd>
                </div>
                <div>
                  <dt className="font-medium">Episode</dt>
                  <dd className="truncate" title={episodeStatus.title || episodeStatus.url}>
                    {episodeStatus.title || episodeStatus.url}
                  </dd>
                </div>
                <div>
                  <dt className="font-medium">Episode ID</dt>
                  <dd className="break-all">{episodeStatus.podcastId}</dd>
                </div>
                <div>
                  <dt className="font-medium">Latest job</dt>
                  <dd>{latestJob?.jobId ? latestJob.jobId : 'None recorded'}</dd>
                </div>
                <div>
                  <dt className="font-medium">Completion</dt>
                  <dd>{episodeStatus.hasTranscript || displayStatus === 'succeeded' ? 'Complete' : 'Not complete'}</dd>
                </div>
                <div>
                  <dt className="font-medium">Source URL</dt>
                  <dd className="break-all">{episodeStatus.url || 'Unknown'}</dd>
                </div>
              </dl>
            )}

            {latestJob?.errorMessage && (
              <div role="alert" className="rounded border border-red-300 bg-white/70 px-3 py-2 text-sm text-red-800">
                {latestJob.errorMessage}
              </div>
            )}

            {requestState === 'error' && error && (
              <div role="alert" className="rounded border border-red-300 bg-red-50 px-3 py-2 text-sm text-red-800">
                {error}
              </div>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}
