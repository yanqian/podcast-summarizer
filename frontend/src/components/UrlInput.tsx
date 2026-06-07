import { useState } from 'react';

type Props = {
  onSubmit: (url: string) => void;
  status?: 'idle' | 'loading' | 'streaming' | 'error' | 'success';
  errorMessage?: string;
};

export function UrlInput({ onSubmit, status = 'idle', errorMessage }: Props) {
  const [value, setValue] = useState('');
  const disabled = status === 'loading' || status === 'streaming';

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!value.trim()) return;
    onSubmit(value.trim());
  };

  return (
    <form className="flex gap-2" onSubmit={handleSubmit} aria-label="url-form">
      <input
        aria-label="podcast-url"
        className="flex-1 rounded border px-3 py-2"
        placeholder="https://example.com/podcast"
        value={value}
        onChange={(e) => setValue(e.target.value)}
        disabled={disabled}
      />
      <button
        type="submit"
        className="rounded bg-blue-600 px-4 py-2 text-white disabled:opacity-50"
        disabled={disabled || !value.trim()}
      >
        {status === 'loading' || status === 'streaming' ? 'Processing...' : 'Submit'}
      </button>
      {status === 'error' && errorMessage ? (
        <span role="alert" className="text-red-600">
          {errorMessage}
        </span>
      ) : null}
    </form>
  );
}
