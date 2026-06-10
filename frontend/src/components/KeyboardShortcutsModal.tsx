'use client';

import { useEffect, useRef } from 'react';
import { X, Keyboard } from 'lucide-react';
import { cn } from '@/lib/utils';

interface KeyboardShortcutsModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const shortcuts = [
  { keys: ['/'], description: 'Focus search input' },
  { keys: ['Escape'], description: 'Close modals / Clear search' },
  { keys: ['j'], description: 'Move down to next incident' },
  { keys: ['k'], description: 'Move up to previous incident' },
  { keys: ['Enter'], description: 'Open selected incident' },
  { keys: ['g', 'h'], description: 'Go to Home' },
  { keys: ['g', 's'], description: 'Go to Settings' },
  { keys: ['?'], description: 'Show this help' },
];

export function KeyboardShortcutsModal({ isOpen, onClose }: KeyboardShortcutsModalProps) {
  const modalRef = useRef<HTMLDivElement>(null);
  const closeButtonRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (isOpen) {
      closeButtonRef.current?.focus();
    }
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      role="dialog"
      aria-modal="true"
      aria-label="Keyboard shortcuts"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div
        ref={modalRef}
        className="bg-zinc-900 border border-zinc-800 rounded-xl p-6 w-full max-w-md mx-4 shadow-2xl"
      >
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-3">
            <Keyboard className="w-5 h-5 text-blue-500" aria-hidden="true" />
            <h2 className="text-lg font-semibold text-zinc-100">Keyboard Shortcuts</h2>
          </div>
          <button
            ref={closeButtonRef}
            onClick={onClose}
            className="p-1 text-zinc-500 hover:text-zinc-300 transition-colors rounded"
            aria-label="Close shortcuts modal"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="space-y-3">
          {shortcuts.map((shortcut, index) => (
            <div key={index} className="flex items-center justify-between py-2 border-b border-zinc-800 last:border-0">
              <span className="text-sm text-zinc-400">{shortcut.description}</span>
              <div className="flex items-center gap-1">
                {shortcut.keys.map((key, keyIndex) => (
                  <span key={keyIndex} className="flex items-center gap-1">
                    {keyIndex > 0 && <span className="text-xs text-zinc-600">then</span>}
                    <kbd
                      className={cn(
                        'inline-flex items-center justify-center min-w-[24px] h-6 px-2',
                        'bg-zinc-800 border border-zinc-700 rounded text-xs font-mono',
                        'text-zinc-300'
                      )}
                    >
                      {key}
                    </kbd>
                  </span>
                ))}
              </div>
            </div>
          ))}
        </div>

        <div className="mt-6 pt-4 border-t border-zinc-800">
          <p className="text-xs text-zinc-500 text-center">
            Press <kbd className="px-1.5 py-0.5 bg-zinc-800 border border-zinc-700 rounded text-xs font-mono text-zinc-300">?</kbd> anywhere to toggle this help
          </p>
        </div>
      </div>
    </div>
  );
}
