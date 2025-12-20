import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/react';
import { UrlInput } from '../../src/components/UrlInput';

describe('UrlInput', () => {
  it('submits valid URL', () => {
    const onSubmit = vi.fn();
    render(<UrlInput onSubmit={onSubmit} status="idle" />);

    const input = screen.getByLabelText('podcast-url');
    fireEvent.change(input, { target: { value: 'http://example.com/podcast' } });
    fireEvent.submit(screen.getByLabelText('url-form'));

    expect(onSubmit).toHaveBeenCalledWith('http://example.com/podcast');
  });

  it('shows error message when provided', () => {
    render(<UrlInput onSubmit={() => {}} status="error" errorMessage="Invalid URL" />);
    expect(screen.getByRole('alert').textContent).toContain('Invalid URL');
  });
});
