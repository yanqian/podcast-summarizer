import { useState } from 'react';
import { PodcastList } from '../components/PodcastList';
import { PodcastListItem } from '../services/podcastClient';
import { PodcastView } from './PodcastView';

const App = () => {
  const [selected, setSelected] = useState<PodcastListItem | undefined>();
  const [refreshKey, setRefreshKey] = useState(0);

  const handleSelect = (item: PodcastListItem) => {
    setSelected(item);
  };

  const handleIngested = (podcastId: string, jobId: string) => {
    setSelected({ id: podcastId, url: '', title: '', hasTranscript: false, createdAt: new Date().toISOString(), latestJobId: jobId });
    setRefreshKey((k) => k + 1);
  };

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900">
      <div className="mx-auto max-w-6xl px-6 py-12 space-y-6">
        <header>
          <h1 className="text-3xl font-bold">Podcast Summary</h1>
          <p className="text-slate-600">Paste a podcast URL to view transcript and aligned summaries.</p>
        </header>
        <div className="grid gap-6 md:grid-cols-3">
          <div className="md:col-span-1">
            <PodcastList selectedId={selected?.id} onSelect={handleSelect} refreshKey={refreshKey} />
          </div>
          <div className="md:col-span-2">
            <PodcastView selectedPodcastId={selected?.id} onIngested={handleIngested} />
          </div>
        </div>
      </div>
    </div>
  );
};

export default App;
