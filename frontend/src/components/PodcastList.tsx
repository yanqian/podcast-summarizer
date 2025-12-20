import { useEffect, useState } from 'react';
import { fetchPodcasts, PodcastListItem } from '../services/podcastClient';

type Props = {
  selectedId?: string;
  onSelect: (item: PodcastListItem) => void;
  refreshKey?: number;
};

export function PodcastList({ selectedId, onSelect, refreshKey }: Props) {
  const [items, setItems] = useState<PodcastListItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | undefined>();

  useEffect(() => {
    setLoading(true);
    setError(undefined);
    fetchPodcasts()
      .then(setItems)
      .catch((e) => setError(e instanceof Error ? e.message : 'Failed to load podcasts'))
      .finally(() => setLoading(false));
  }, [refreshKey]);

  return (
    <section className="rounded border bg-white p-4 shadow-sm">
      <div className="mb-3 flex items-center justify-between">
        <h2 className="text-lg font-semibold">Recent Podcasts</h2>
        {loading && <span className="text-sm text-slate-500">Loading…</span>}
      </div>
      {error && <p className="mb-2 text-sm text-red-600">{error}</p>}
      <div className="divide-y">
        {items.map((item) => (
          <button
            key={item.id}
            onClick={() => onSelect(item)}
            className={`flex w-full items-start gap-3 py-3 text-left transition hover:bg-slate-50 ${
              selectedId === item.id ? 'bg-slate-100' : ''
            }`}
          >
            <div className="min-w-0 flex-1 space-y-1">
              <p className="text-sm font-semibold text-slate-900 truncate" title={item.title || item.url}>
                {item.title || item.url}
              </p>
              <p className="text-[11px] text-slate-600 break-all line-clamp-1" title={item.url}>
                {item.url}
              </p>
              <p className="text-[11px] text-slate-500">
                Added: {item.createdAt ? new Date(item.createdAt).toLocaleString() : 'Unknown'}
              </p>
            </div>
            <div className="text-right text-xs text-slate-600 min-w-[96px]">
              <div
                className={`inline-block rounded-full px-2 py-[2px] text-[11px] font-semibold ${
                  (item.latestStatus || '').toLowerCase() === 'succeeded' || item.hasTranscript
                    ? 'bg-emerald-50 text-emerald-700'
                    : (item.latestStatus || '').toLowerCase() === 'failed'
                    ? 'bg-red-50 text-red-700'
                    : 'bg-amber-50 text-amber-700'
                }`}
              >
                {item.latestStatus
                  ? item.latestStatus
                  : item.hasTranscript
                  ? 'ready'
                  : 'pending'}
              </div>
              {item.latestJobId && (
                <div className="mt-1 text-[10px] text-slate-500">Job: {item.latestJobId.slice(0, 8)}…</div>
              )}
            </div>
          </button>
        ))}
        {!loading && items.length === 0 && <p className="py-2 text-sm text-slate-600">No podcasts yet.</p>}
      </div>
    </section>
  );
}
