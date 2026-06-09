import { afterEach, describe, expect, it, vi } from 'vitest';
import { cleanup, fireEvent, render, screen } from '@testing-library/react';
import App from '../../src/pages/App';

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

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe('App', () => {
  it('opens on the Demo screen and switches to Admin without routing', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse({ items: [] })));

    render(<App />);

    expect(screen.getByRole('button', { name: /demo/i }).getAttribute('aria-pressed')).toBe('true');
    expect(screen.getByText('Demo transcript and summary')).toBeTruthy();
    expect(screen.queryByText('Admin submission')).toBeNull();

    fireEvent.click(screen.getByRole('button', { name: /admin/i }));

    expect(screen.getByRole('button', { name: /admin/i }).getAttribute('aria-pressed')).toBe('true');
    expect(screen.getByText('Admin submission')).toBeTruthy();
    expect(screen.queryByText('Demo transcript and summary')).toBeNull();
  });
});
