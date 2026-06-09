import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import { PodcastView } from '../../src/pages/PodcastView';

type FetchResponse = {
  ok: boolean;
  status?: number;
  json?: () => Promise<unknown>;
  text?: () => Promise<string>;
};

const selectedPodcast = {
  id: 'pod-1',
  url: 'https://example.com/feed',
  title: 'Mapping Show',
  hasTranscript: true,
  createdAt: '2026-06-08T00:00:00Z',
  latestStatus: 'succeeded'
};

const jsonResponse = (body: unknown, ok = true): FetchResponse => ({
  ok,
  status: ok ? 200 : 500,
  json: async () => body,
  text: async () => (typeof body === 'string' ? body : JSON.stringify(body))
});

const mockFetch = (responses: FetchResponse[]) => {
  const fetchMock = vi.fn();
  responses.forEach((response) => {
    fetchMock.mockResolvedValueOnce(response);
  });
  vi.stubGlobal('fetch', fetchMock);
  return fetchMock;
};

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('PodcastView', () => {
  it('prompts for an episode before one is selected', () => {
    render(<PodcastView />);

    expect(screen.getByText('Select an episode to view transcript segments and mapped summaries.')).toBeTruthy();
  });

  it('fetches episode detail and renders transcript sections beside mapped summaries', async () => {
    const fetchMock = mockFetch([
      jsonResponse({
        podcastId: 'pod-1',
        id: 'pod-1',
        url: 'https://example.com/feed',
        title: 'Mapping Show',
        hasTranscript: true,
        status: 'succeeded',
        latestJob: { jobId: 'job-1', podcastId: 'pod-1', type: 'ingest', status: 'succeeded' },
        transcriptSegments: [
          { id: 'tr-1', orderIndex: 1, text: 'The host introduces local podcast processing.', startSeconds: 0, endSeconds: 12 },
          { id: 'tr-2', orderIndex: 2, text: 'The guest explains grouped summaries.', startSeconds: 12, endSeconds: 28 },
          { id: 'tr-3', orderIndex: 3, text: 'An unmapped closing segment remains readable.' }
        ],
        summarySegments: [
          {
            id: 'sum-1',
            orderIndex: 1,
            text: 'Intro and grouped summary explanation.',
            sourceTranscriptSegmentIds: ['tr-1', 'tr-2']
          }
        ],
      })
    ]);

    render(<PodcastView selectedPodcast={selectedPodcast} />);

    expect(screen.getByText('Loading episode detail...')).toBeTruthy();
    expect(await screen.findAllByText('Transcript')).toHaveLength(2);

    expect(screen.getByText('The host introduces local podcast processing.')).toBeTruthy();
    expect(screen.getByText('The guest explains grouped summaries.')).toBeTruthy();
    expect(screen.getByText('Intro and grouped summary explanation.')).toBeTruthy();
    expect(screen.getByText('An unmapped closing segment remains readable.')).toBeTruthy();
    expect(screen.getByText('No summary is mapped to this transcript segment yet.')).toBeTruthy();
    expect(fetchMock).toHaveBeenCalledWith(
      'http://localhost:8080/api/podcasts/pod-1',
      expect.objectContaining({ method: 'GET' })
    );
  });

  it('toggles long transcript sections', async () => {
    mockFetch([
      jsonResponse({
        podcastId: 'pod-1',
        id: 'pod-1',
        url: 'https://example.com/feed',
        title: 'Long Show',
        hasTranscript: true,
        status: 'succeeded',
        transcriptSegments: [
          {
            id: 'tr-1',
            orderIndex: 1,
            text: 'Long transcript text. '.repeat(80)
          }
        ],
        summarySegments: [
          {
            id: 'sum-1',
            orderIndex: 1,
            text: 'Short summary.',
            sourceTranscriptSegmentIds: ['tr-1']
          }
        ]
      })
    ]);

    render(<PodcastView selectedPodcast={{ ...selectedPodcast, title: 'Long Show' }} />);

    const showMore = await screen.findByRole('button', { name: 'Show more' });
    fireEvent.click(showMore);
    expect(screen.getByRole('button', { name: 'Show less' })).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Show less' }));
    expect(screen.getByRole('button', { name: 'Show more' })).toBeTruthy();
  });

  it('paginates transcript summary cards ten at a time', async () => {
    const transcriptSegments = Array.from({ length: 11 }, (_, idx) => ({
      id: `tr-${idx + 1}`,
      orderIndex: idx + 1,
      text: `Transcript segment ${idx + 1}.`
    }));
    const summarySegments = transcriptSegments.map((segment) => ({
      id: `sum-${segment.orderIndex}`,
      orderIndex: segment.orderIndex,
      text: `Summary ${segment.orderIndex}.`,
      sourceTranscriptSegmentIds: [segment.id]
    }));
    mockFetch([
      jsonResponse({
        podcastId: 'pod-1',
        id: 'pod-1',
        url: 'https://example.com/feed',
        title: 'Paged Show',
        hasTranscript: true,
        status: 'succeeded',
        transcriptSegments,
        summarySegments
      })
    ]);

    render(<PodcastView selectedPodcast={{ ...selectedPodcast, title: 'Paged Show' }} />);

    await screen.findByText('Page 1 of 2');
    expect(screen.getByText('Segment 10')).toBeTruthy();
    expect(screen.queryByText('Segment 11')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: 'Next' }));
    expect(screen.getByText('Page 2 of 2')).toBeTruthy();
    expect(screen.getByText('Segment 11')).toBeTruthy();
    expect(screen.queryByText('Segment 1')).toBeNull();
  });

  it('shows an empty completed episode state when no transcript segments exist', async () => {
    mockFetch([
      jsonResponse({
        podcastId: 'pod-1',
        id: 'pod-1',
        url: 'https://example.com/feed',
        title: 'Empty Show',
        hasTranscript: false,
        status: 'not_started',
        transcriptSegments: [],
        summarySegments: []
      })
    ]);

    render(<PodcastView selectedPodcast={{ ...selectedPodcast, title: 'Empty Show', hasTranscript: false }} />);

    await screen.findByText('No transcript segments available yet.');
  });

  it('shows failed processing state and latest job error', async () => {
    mockFetch([
      jsonResponse({
        podcastId: 'pod-1',
        id: 'pod-1',
        url: 'https://example.com/feed',
        title: 'Failed Show',
        hasTranscript: false,
        status: 'failed',
        latestJob: {
          jobId: 'job-1',
          podcastId: 'pod-1',
          type: 'ingest',
          status: 'failed',
          errorMessage: 'chunk audio failed'
        },
        transcriptSegments: [],
        summarySegments: []
      })
    ]);

    render(<PodcastView selectedPodcast={{ ...selectedPodcast, title: 'Failed Show', latestStatus: 'failed' }} />);

    await screen.findByText('Processing failed');
    expect(screen.getByText('chunk audio failed')).toBeTruthy();
  });
});
