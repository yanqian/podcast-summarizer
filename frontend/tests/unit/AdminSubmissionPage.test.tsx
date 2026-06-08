import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { AdminSubmissionPage } from '../../src/pages/AdminSubmissionPage';

type FetchResponse = {
  ok: boolean;
  status?: number;
  json?: () => Promise<unknown>;
  text?: () => Promise<string>;
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

describe('AdminSubmissionPage', () => {
  it('submits a new podcast URL and displays queued status', async () => {
    const onSubmissionRecorded = vi.fn();
    const fetchMock = mockFetch([
      jsonResponse({ podcastId: 'pod-1', jobId: 'job-1', status: 'queued', existing: false }),
      jsonResponse({
        podcastId: 'pod-1',
        id: 'pod-1',
        url: 'https://example.com/feed',
        title: 'Example Show',
        hasTranscript: false,
        status: 'queued',
        latestJob: { jobId: 'job-1', podcastId: 'pod-1', type: 'ingest', status: 'queued' }
      })
    ]);

    render(<AdminSubmissionPage onSubmissionRecorded={onSubmissionRecorded} />);

    fireEvent.change(screen.getByLabelText('podcast-url'), { target: { value: 'https://example.com/feed' } });
    fireEvent.submit(screen.getByLabelText('url-form'));

    expect(screen.getByText('Processing...')).toBeTruthy();
    await screen.findByText('New processing job started');
    expect(screen.getByText('queued')).toBeTruthy();
    expect(screen.getByText('job-1')).toBeTruthy();
    expect(onSubmissionRecorded).toHaveBeenCalledTimes(1);
    expect(fetchMock).toHaveBeenNthCalledWith(
      1,
      'http://localhost:8080/api/podcasts/ingest',
      expect.objectContaining({ method: 'POST' })
    );
  });

  it('shows existing episode state instead of implying a new run', async () => {
    mockFetch([
      jsonResponse({ podcastId: 'pod-2', jobId: 'job-2', status: 'existing', existing: true, latestStatus: 'succeeded' }),
      jsonResponse({
        podcastId: 'pod-2',
        id: 'pod-2',
        url: 'https://example.com/existing',
        title: 'Existing Show',
        hasTranscript: true,
        status: 'succeeded',
        latestJob: { jobId: 'job-2', podcastId: 'pod-2', type: 'ingest', status: 'succeeded' }
      })
    ]);

    render(<AdminSubmissionPage />);

    fireEvent.change(screen.getByLabelText('podcast-url'), { target: { value: 'https://example.com/existing' } });
    fireEvent.submit(screen.getByLabelText('url-form'));

    await screen.findByText('Existing episode found');
    expect(screen.getByText(/no new processing run was implied/i)).toBeTruthy();
    expect(screen.getByText('Complete')).toBeTruthy();
    expect(screen.getByText('succeeded')).toBeTruthy();
  });

  it('loads selected failed episode status and displays the processing error', async () => {
    mockFetch([
      jsonResponse({
        podcastId: 'pod-3',
        id: 'pod-3',
        url: 'https://example.com/failed',
        title: 'Failed Show',
        hasTranscript: false,
        status: 'failed',
        latestJob: {
          jobId: 'job-3',
          podcastId: 'pod-3',
          type: 'ingest',
          status: 'failed',
          errorMessage: 'transcript pipeline failed'
        }
      })
    ]);

    render(
      <AdminSubmissionPage
        selectedPodcast={{
          id: 'pod-3',
          url: 'https://example.com/failed',
          title: 'Failed Show',
          hasTranscript: false,
          createdAt: '2026-06-08T00:00:00Z',
          latestJobId: 'job-3',
          latestStatus: 'failed'
        }}
      />
    );

    await screen.findByText('Episode status');
    expect(screen.getByText('failed')).toBeTruthy();
    expect(screen.getByRole('alert').textContent).toContain('transcript pipeline failed');
  });

  it('shows submit errors from the backend', async () => {
    mockFetch([jsonResponse('Bad Request', false)]);

    render(<AdminSubmissionPage />);

    fireEvent.change(screen.getByLabelText('podcast-url'), { target: { value: 'not-a-valid-feed' } });
    fireEvent.submit(screen.getByLabelText('url-form'));

    await waitFor(() => {
      expect(screen.getAllByRole('alert').map((node) => node.textContent).join(' ')).toContain('Bad Request');
    });
  });
});
