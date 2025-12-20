type JobStatus = {
  jobId: string;
  podcastId: string;
  status: 'queued' | 'running' | 'succeeded' | 'failed';
  errorMessage?: string | null;
};

type PollOptions = {
  intervalMs?: number;
  timeoutMs?: number;
  fetchStatus: (jobId: string) => Promise<JobStatus>;
  onUpdate?: (status: JobStatus) => void;
};

export async function pollJobStatus(jobId: string, options: PollOptions): Promise<JobStatus> {
  const interval = options.intervalMs ?? 1000;
  const timeout = options.timeoutMs ?? 5 * 60 * 1000;
  const start = Date.now();
  let currentInterval = interval;

  while (true) {
    const status = await options.fetchStatus(jobId);
    options.onUpdate?.(status);
    if (status.status === 'succeeded' || status.status === 'failed') {
      return status;
    }
    if (Date.now() - start > timeout) {
      throw new Error('Job polling timed out');
    }
    await new Promise((resolve) => setTimeout(resolve, currentInterval));
    // simple backoff to reduce load
    if (currentInterval < 5000) {
      currentInterval += 500;
    }
  }
}
