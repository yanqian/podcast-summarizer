import { useEffect, useLayoutEffect, useRef, useState } from 'react';
import { UrlInput } from '../components/UrlInput';
import { fetchView, ingestPodcast } from '../services/podcastClient';
import { StatusBanner } from '../components/StatusBanner';
import { subscribeTranscript, StreamEvent } from '../services/streamClient';

type Paragraph = {
  paragraphId: string;
  orderIndex: number;
  text: string;
  summary?: string;
};

type StructuredSummary = {
  discussion: string;
};

function parseStructuredSummary(raw?: string): StructuredSummary | null {
  if (!raw) return null;
  const trimmed = raw.trim();
  if (!trimmed.startsWith('{')) return null;
  try {
    const parsed: unknown = JSON.parse(trimmed);
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return null;
    const record = parsed as Record<string, unknown>;
    const discussion = typeof record.discussion === 'string' ? record.discussion.trim() : '';
    if (!discussion) return null;
    return { discussion };
  } catch {
    return null;
  }
}

type Props = {
  selectedPodcastId?: string;
  onIngested?: (podcastId: string, jobId: string) => void;
};

export const PodcastView = ({ selectedPodcastId, onIngested }: Props) => {
  const [status, setStatus] = useState<'idle' | 'loading' | 'streaming' | 'error' | 'success'>('idle');
  const [error, setError] = useState<string | undefined>();
  const [paragraphs, setParagraphs] = useState<Paragraph[]>([]);
  const [podcastId, setPodcastId] = useState<string | undefined>();
  const [pageIndex, setPageIndex] = useState<number>(0);
  const [selectedIndex, setSelectedIndex] = useState<number | undefined>();
  const transcriptCardRef = useRef<HTMLDivElement | null>(null);
  const [cardHeight, setCardHeight] = useState<number>();

  const loadExisting = (id: string) => {
    setStatus('loading');
    setError(undefined);
    setParagraphs([]);
    fetchView(id)
      .then((view) => {
        const paras = Array.isArray(view.paragraphs) ? view.paragraphs : [];
        const normalized = paras.map((p) => ({
          paragraphId: p.paragraphId,
          orderIndex: p.orderIndex ?? 0,
          text: p.text,
          summary: p.summary
        }));
        setPodcastId(view.podcastId);
        setParagraphs(normalized);
        setStatus('success');
      })
      .catch((e) => {
        setStatus('error');
        setError(e instanceof Error ? e.message : 'Unexpected error');
      });
  };

  // Load when a podcast is selected from list
  useEffect(() => {
    if (selectedPodcastId && selectedPodcastId !== podcastId && status !== 'streaming') {
      loadExisting(selectedPodcastId);
    }
  }, [selectedPodcastId]);

  const handleStreamEvent = (evt: StreamEvent) => {
    if (evt.event === 'chunk') {
      setParagraphs((prev) => {
        const next = [...prev];
        next.push({
          paragraphId: `stream-${evt.data.order}`,
          orderIndex: evt.data.order,
          text: evt.data.text
        });
        return next;
      });
    } else if (evt.event === 'done' && podcastId) {
      fetchView(podcastId)
        .then((view) => {
          const paras = Array.isArray(view.paragraphs) ? view.paragraphs : [];
          const normalized = paras.map((p) => ({
            paragraphId: p.paragraphId,
            orderIndex: p.orderIndex ?? 0,
            text: p.text,
            summary: p.summary
          }));
          setParagraphs(normalized);
          setStatus('success');
        })
        .catch((e) => {
          setStatus('error');
          setError(e instanceof Error ? e.message : 'Unexpected error');
        });
    }
  };

  const handleSubmit = async (url: string) => {
    setStatus('loading');
    setError(undefined);
    setParagraphs([]);
    try {
      const job = await ingestPodcast(url);
      setPodcastId(job.podcastId);
      onIngested?.(job.podcastId, job.jobId);
      setStatus('streaming');
      subscribeTranscript(job.jobId, handleStreamEvent, (err) => {
        setStatus('error');
        setError(err instanceof Event ? 'Streaming error' : 'Unexpected error');
      });
    } catch (e) {
      setStatus('error');
      setError(e instanceof Error ? e.message : 'Unexpected error');
    }
  };

  const currentParagraph = paragraphs[pageIndex];
  const selectedParagraph = selectedIndex !== undefined ? paragraphs.find((p) => p.orderIndex === selectedIndex) : undefined;
  const selectedSummary = parseStructuredSummary(selectedParagraph?.summary);

  const handlePrev = () => {
    setSelectedIndex(undefined);
    setPageIndex((idx) => Math.max(0, idx - 1));
  };
  const handleNext = () => {
    setSelectedIndex(undefined);
    setPageIndex((idx) => Math.min(paragraphs.length - 1, idx + 1));
  };

  useEffect(() => {
    if (paragraphs.length > 0) {
      setPageIndex((idx) => Math.min(idx, paragraphs.length - 1));
      setSelectedIndex(paragraphs[Math.min(pageIndex, paragraphs.length - 1)].orderIndex);
    } else {
      setPageIndex(0);
      setSelectedIndex(undefined);
    }
  }, [paragraphs.length]);

  useEffect(() => {
    if (currentParagraph) {
      setSelectedIndex(currentParagraph.orderIndex);
    }
  }, [pageIndex]);

  // Keep summary card height locked to transcript card height for consistent columns.
  const syncCardHeight = () => {
    const el = transcriptCardRef.current;
    if (!el) return;
    setCardHeight(el.getBoundingClientRect().height);
  };

  useLayoutEffect(() => {
    syncCardHeight();
  }, [paragraphs, selectedIndex]);

  useEffect(() => {
    const onResize = () => syncCardHeight();
    window.addEventListener('resize', onResize);
    return () => window.removeEventListener('resize', onResize);
  }, []);

  return (
    <div className="space-y-6">
      <UrlInput onSubmit={handleSubmit} status={status} errorMessage={error} />
      <StatusBanner status={status} error={error} />

      {paragraphs.length > 1 && (
        <div className="flex items-center justify-center gap-3">
          <button
            onClick={handlePrev}
            disabled={pageIndex === 0}
            className="rounded border px-3 py-1 text-sm disabled:opacity-50"
          >
            Previous
          </button>
          <span className="text-sm text-slate-600">
            Page {pageIndex + 1} of {paragraphs.length}
          </span>
          <button
            onClick={handleNext}
            disabled={pageIndex >= paragraphs.length - 1}
            className="rounded border px-3 py-1 text-sm disabled:opacity-50"
          >
            Next
          </button>
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 items-stretch">
        <section
          ref={transcriptCardRef}
          className="flex h-[70vh] flex-col rounded border bg-white p-4 shadow-sm"
        >
          <h2 className="mb-2 text-lg font-semibold">Transcript</h2>
          <div className="flex-1 space-y-2 overflow-auto pr-1">
            {currentParagraph ? (
              <button
                onClick={() => {
                  setSelectedIndex(currentParagraph.orderIndex);
                }}
                className={`w-full rounded border px-3 py-3 text-left text-sm transition hover:border-slate-300 hover:bg-slate-50 ${
                  selectedIndex === currentParagraph.orderIndex ? 'border-indigo-400 bg-indigo-50' : 'border-transparent'
                }`}
              >
                <span className="text-slate-800 whitespace-pre-line break-words">{currentParagraph.text}</span>
              </button>
            ) : (
              <p className="text-sm text-slate-600">Transcript paragraphs will appear here.</p>
            )}
          </div>
        </section>
        <section
          className="flex h-[70vh] flex-col rounded border bg-white p-4 shadow-sm md:sticky md:top-2"
          style={cardHeight ? { height: cardHeight } : undefined}
        >
          <h2 className="mb-2 text-lg font-semibold">Summary</h2>
          {selectedIndex !== undefined ? (
            <div className="flex-1 rounded border border-indigo-200 bg-slate-50 px-3 py-3 text-sm text-slate-800 shadow-inner">
              {selectedSummary ? (
                <div className="whitespace-pre-line break-words">{selectedSummary.discussion}</div>
              ) : (
                <div className="whitespace-pre-line break-words">{selectedParagraph?.summary || '...'}</div>
              )}
            </div>
          ) : (
            <p className="text-sm text-slate-600">Summaries will align with transcript paragraphs.</p>
          )}
        </section>
      </div>
    </div>
  );
};
