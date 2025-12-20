type Props = {
  status: 'queued' | 'running' | 'succeeded' | 'failed';
  message?: string;
};

export function ProgressIndicator({ status, message }: Props) {
  const base = "rounded border p-2 text-sm";
  if (status === 'queued' || status === 'running') {
    return <div className={`${base} border-blue-200 bg-blue-50 text-blue-700`}>{message || 'Processing...'}</div>;
  }
  if (status === 'failed') {
    return <div className={`${base} border-red-200 bg-red-50 text-red-700`}>{message || 'Failed'}</div>;
  }
  return <div className={`${base} border-green-200 bg-green-50 text-green-700`}>{message || 'Completed'}</div>;
}
