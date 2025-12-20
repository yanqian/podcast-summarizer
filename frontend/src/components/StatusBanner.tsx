type Props = {
  status: 'idle' | 'loading' | 'streaming' | 'error' | 'success';
  error?: string;
};

export function StatusBanner({ status, error }: Props) {
  if (status === 'idle') return null;
  if (status === 'loading') {
    return <div className="rounded border border-blue-200 bg-blue-50 p-2 text-blue-700">Processing...</div>;
  }
  if (status === 'streaming') {
    return <div className="rounded border border-blue-200 bg-blue-50 p-2 text-blue-700">Streaming transcript...</div>;
  }
  if (status === 'error') {
    return (
      <div className="rounded border border-red-200 bg-red-50 p-2 text-red-700" role="alert">
        {error || 'Something went wrong'}
      </div>
    );
  }
  return <div className="rounded border border-green-200 bg-green-50 p-2 text-green-700">Completed</div>;
}
