'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { getSettings, updateSettings, getMe, clearToken } from '@/lib/api';

export default function SettingsPage() {
  const router = useRouter();
  const [openrouterKey, setOpenrouterKey] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [email, setEmail] = useState('');
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const loadData = async () => {
      try {
        const [settingsRes, meRes] = await Promise.all([getSettings(), getMe()]);
        setOpenrouterKey(settingsRes.openrouter_key || '');
        setApiKey(settingsRes.api_key || '');
        setEmail(meRes.user.email);
      } catch {
        router.push('/login');
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, [router]);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError('');
    setMessage('');

    try {
      await updateSettings(openrouterKey);
      setMessage('Settings saved successfully');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  const handleLogout = () => {
    clearToken();
    localStorage.removeItem('nexus_user');
    router.push('/login');
  };

  const copyApiKey = () => {
    navigator.clipboard.writeText(apiKey);
    setMessage('API key copied to clipboard');
  };

  if (loading) {
    return (
      <main className="min-h-screen bg-background flex items-center justify-center">
        <div className="text-muted-foreground" role="status" aria-label="Loading settings">Loading...</div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-background p-4 md:p-8">
      <div className="max-w-2xl mx-auto">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-2xl font-bold text-white">Settings</h1>
            <p className="text-muted-foreground mt-1">{email}</p>
          </div>
          <button
            onClick={handleLogout}
            className="px-4 py-2 bg-surface hover:bg-surface-hover text-foreground rounded transition-colors text-sm"
          >
            Sign Out
          </button>
        </div>

        <div className="space-y-6">
          <section className="bg-card rounded-lg p-6 border border-border">
            <h2 className="text-lg font-semibold text-white mb-4">API Access</h2>
            <p className="text-sm text-muted-foreground mb-4">
              Use this API key to send events to Nexus from your SIEM or other tools.
            </p>
            <div className="flex items-center gap-2">
              <code className="flex-1 px-3 py-2 bg-surface rounded text-success text-sm font-mono break-all">
                {apiKey}
              </code>
              <button
                onClick={copyApiKey}
                className="px-3 py-2 bg-surface-hover hover:bg-surface-active text-foreground rounded text-sm whitespace-nowrap"
              >
                Copy
              </button>
            </div>
          </section>

          <section className="bg-card rounded-lg p-6 border border-border">
            <h2 className="text-lg font-semibold text-white mb-4">OpenRouter API Key</h2>
            <p className="text-sm text-muted-foreground mb-4">
              Provide your own OpenRouter API key to generate narratives. Get a free key at{' '}
              <a
                href="https://openrouter.ai/keys"
                target="_blank"
                rel="noopener noreferrer"
                className="text-primary hover:text-primary/80"
              >
                openrouter.ai/keys
              </a>
            </p>

            <form onSubmit={handleSave}>
              {message && (
                <div className="mb-4 p-3 bg-success/10 border border-success/30 rounded text-success text-sm" role="status">
                  {message}
                </div>
              )}
              {error && (
                <div className="mb-4 p-3 bg-destructive/10 border border-destructive/30 rounded text-destructive text-sm" role="alert">
                  {error}
                </div>
              )}

              <div className="mb-4">
                <input
                  type="password"
                  value={openrouterKey}
                  onChange={(e) => setOpenrouterKey(e.target.value)}
                  placeholder="sk-or-v1-..."
                  className="w-full px-3 py-2 bg-surface border border-border rounded text-white font-mono text-sm focus:outline-none focus:ring-2 focus:ring-primary"
                  aria-label="OpenRouter API Key"
                />
              </div>

              <button
                type="submit"
                disabled={saving}
                className="w-full py-2 px-4 bg-primary hover:bg-primary/90 disabled:bg-primary/50 text-white font-medium rounded transition-colors"
              >
                {saving ? 'Saving...' : 'Save Settings'}
              </button>
            </form>
          </section>
        </div>
      </div>
    </main>
  );
}
