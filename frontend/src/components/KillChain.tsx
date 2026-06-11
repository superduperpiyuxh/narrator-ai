'use client';

import { ChevronRight } from 'lucide-react';
import type { TechniqueRef } from '@/lib/types';

interface KillChainProps {
  techniques: TechniqueRef[];
}

interface KillChainPhase {
  label: string;
  tactic: string | null;
}

const KILL_CHAIN_PHASES: KillChainPhase[] = [
  { label: 'Reconnaissance', tactic: null },
  { label: 'Resource Development', tactic: null },
  { label: 'Initial Access', tactic: 'initial-access' },
  { label: 'Execution', tactic: 'execution' },
  { label: 'Persistence', tactic: 'persistence' },
  { label: 'Privilege Escalation', tactic: 'privilege-escalation' },
  { label: 'Defense Evasion', tactic: 'defense-evasion' },
  { label: 'Credential Access', tactic: 'credential-access' },
  { label: 'Discovery', tactic: 'discovery' },
  { label: 'Lateral Movement', tactic: 'lateral-movement' },
  { label: 'Collection', tactic: 'collection' },
  { label: 'Command and Control', tactic: 'command-and-control' },
  { label: 'Exfiltration', tactic: 'exfiltration' },
  { label: 'Impact', tactic: 'impact' },
];

export function KillChain({ techniques }: KillChainProps) {
  // Build a map from tactic slug -> list of technique IDs
  const tacticToTechniques = new Map<string, string[]>();
  for (const tech of techniques) {
    const tactic = tech.tactic;
    if (!tacticToTechniques.has(tactic)) {
      tacticToTechniques.set(tactic, []);
    }
    tacticToTechniques.get(tactic)!.push(tech.technique_id);
  }

  // Determine which phase indices are active
  const activeIndices = new Set<number>();
  KILL_CHAIN_PHASES.forEach((phase, idx) => {
    if (phase.tactic && tacticToTechniques.has(phase.tactic)) {
      activeIndices.add(idx);
    }
  });

  return (
    <div className="overflow-x-auto pb-2 -mx-1 px-1">
      <div className="flex items-center gap-0 min-w-max">
        {KILL_CHAIN_PHASES.map((phase, idx) => {
          const isActive = activeIndices.has(idx);
          const techIds = phase.tactic ? (tacticToTechniques.get(phase.tactic) ?? []) : [];
          const isLast = idx === KILL_CHAIN_PHASES.length - 1;

          return (
            <div key={phase.label} className="flex items-center">
              {/* Phase box */}
              <div
                className={`
                  flex flex-col items-center justify-center rounded-lg border px-3 py-2 min-w-[110px] max-w-[140px] text-center
                  transition-colors
                  ${isActive
                    ? 'bg-primary/20 border-primary text-primary/80'
                    : 'bg-card border-border text-muted-foreground/60'
                  }
                `}
              >
                <span className="text-[11px] font-medium leading-tight whitespace-nowrap">
                  {phase.label}
                </span>
                {isActive && techIds.length > 0 && (
                  <span className="mt-1 text-[10px] font-mono text-primary/80 leading-tight">
                    {techIds.join(', ')}
                  </span>
                )}
              </div>

              {/* Arrow between phases */}
              {!isLast && (
                <ChevronRight
                  className={`w-4 h-4 flex-shrink-0 mx-0.5 ${
                    isActive ? 'text-primary' : 'text-muted-foreground/60'
                  }`}
                  aria-hidden="true"
                />
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
