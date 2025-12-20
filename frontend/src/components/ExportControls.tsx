type Props = {
  transcript: string;
  summary: string;
};

export function ExportControls({ transcript, summary }: Props) {
  const download = (text: string, filename: string) => {
    const blob = new Blob([text], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  };

  const copy = async (text: string) => {
    await navigator.clipboard.writeText(text);
    alert('Copied to clipboard');
  };

  return (
    <div className="flex gap-2">
      <button
        type="button"
        className="rounded border px-3 py-2 text-sm"
        onClick={() => download(transcript, 'transcript.txt')}
      >
        Export Transcript
      </button>
      <button
        type="button"
        className="rounded border px-3 py-2 text-sm"
        onClick={() => download(summary, 'summary.txt')}
      >
        Export Summary
      </button>
      <button
        type="button"
        className="rounded border px-3 py-2 text-sm"
        onClick={() => copy(`${transcript}\n\n${summary}`)}
      >
        Copy Both
      </button>
    </div>
  );
}
