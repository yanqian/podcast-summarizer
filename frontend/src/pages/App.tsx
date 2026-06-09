import { useState } from 'react';
import { Settings, Tv } from 'lucide-react';
import { PodcastList } from '../components/PodcastList';
import { PodcastListItem } from '../services/podcastClient';
import { AdminSubmissionPage } from './AdminSubmissionPage';
import { PodcastView } from './PodcastView';

type Screen = 'demo' | 'admin';

const App = () => {
  const [selected, setSelected] = useState<PodcastListItem | undefined>();
  const [refreshKey, setRefreshKey] = useState(0);
  const [screen, setScreen] = useState<Screen>('demo');

  const handleSelect = (item: PodcastListItem) => {
    setSelected(item);
  };

  const handleSubmissionRecorded = () => {
    setRefreshKey((k) => k + 1);
  };

  return (
    <div className="min-h-screen bg-zinc-50 text-zinc-950">
      <div className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
        <header className="mb-6 flex flex-col gap-4 border-b border-zinc-200 pb-5 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h1 className="text-3xl font-bold">Podcast Summarizer Demo</h1>
            <p className="mt-1 text-sm text-zinc-600">
              Browse local episode summaries or manage podcast ingestion from the admin screen.
            </p>
          </div>
          <nav aria-label="Primary screens" className="inline-flex w-full rounded border border-zinc-300 bg-white p-1 sm:w-auto">
            <button
              type="button"
              onClick={() => setScreen('demo')}
              className={`inline-flex flex-1 items-center justify-center gap-2 rounded px-3 py-2 text-sm font-medium sm:flex-none ${
                screen === 'demo' ? 'bg-zinc-900 text-white' : 'text-zinc-700 hover:bg-zinc-100'
              }`}
              aria-pressed={screen === 'demo'}
            >
              <Tv aria-hidden="true" className="h-4 w-4" />
              Demo
            </button>
            <button
              type="button"
              onClick={() => setScreen('admin')}
              className={`inline-flex flex-1 items-center justify-center gap-2 rounded px-3 py-2 text-sm font-medium sm:flex-none ${
                screen === 'admin' ? 'bg-zinc-900 text-white' : 'text-zinc-700 hover:bg-zinc-100'
              }`}
              aria-pressed={screen === 'admin'}
            >
              <Settings aria-hidden="true" className="h-4 w-4" />
              Admin
            </button>
          </nav>
        </header>

        <div className="grid gap-6 md:grid-cols-[320px_minmax(0,1fr)]">
          <div className="md:col-span-1">
            <PodcastList selectedId={selected?.id} onSelect={handleSelect} refreshKey={refreshKey} />
          </div>
          <main className="min-w-0">
            {screen === 'admin' ? (
              <AdminSubmissionPage selectedPodcast={selected} onSubmissionRecorded={handleSubmissionRecorded} />
            ) : (
              <PodcastView selectedPodcast={selected} />
            )}
          </main>
        </div>
      </div>
    </div>
  );
};

export default App;
