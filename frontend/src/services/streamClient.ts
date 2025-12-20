export type StreamEvent =
  | { event: 'chunk'; data: { order: number; text: string } }
  | { event: 'done'; data: Record<string, unknown> };

export function subscribeTranscript(
  jobId: string,
  onEvent: (evt: StreamEvent) => void,
  onError?: (err: Event) => void
): EventSource {
  const base = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';
  const url = `${base}/api/streams/transcript/${jobId}`;
  const es = new EventSource(url);
  es.onmessage = (ev) => {
    try {
      const parsed = JSON.parse(ev.data);
      onEvent(parsed);
    } catch {
      onError?.(new Event('parse-error'));
    }
  };
  if (onError) {
    es.onerror = (err) => onError(err);
  }
  return es;
}
