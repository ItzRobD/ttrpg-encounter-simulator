import { Injectable } from '@angular/core';
import { SimulationEvent, TimelineNode } from '../models';

@Injectable({
  providedIn: 'root',
})
export class Mapper {
  /**
   * Transforms a raw gzipped JSON log into a hierarchical tree for the UI.
   */
  mapSimulationLog(rawLog: unknown[]): TimelineNode[] {
    const events = rawLog.map((e) => this.mapKeys(e) as SimulationEvent);
    return this.buildTree(events);
  }

  /**
   * Recursively converts object keys from PascalCase to camelCase.
   * Uses 'unknown' because JSON data can contain heterogeneous types (strings, numbers, booleans, nested objects).
   * Unlike 'any', 'unknown' is type-safe as it forces checking or casting before the data can be used.
   */
  mapKeys(obj: unknown): unknown {
    if (Array.isArray(obj)) {
      return obj.map((v) => this.mapKeys(v));
    } else if (obj !== null && typeof obj === 'object') {
      const result: Record<string, unknown> = {};
      const record = obj as Record<string, unknown>;

      Object.keys(record).forEach((key) => {
        const camelKey = this.toCamelCase(key);
        result[camelKey] = this.mapKeys(record[key]);
      });

      return result;
    }
    return obj;
  }

  /**
   * Robust PascalCase to camelCase converter that handles dnd specific abbreviations.
   */
  private toCamelCase(str: string): string {
    if (str === 'ID') return 'id';
    if (str === 'AC') return 'ac';
    if (str === 'HP') return 'hp';

    let result = str;

    // Handle common acronyms at the end of strings (e.g., SequenceID -> SequenceId, MaxHP -> MaxHp)
    if (result.endsWith('ID')) {
      result = result.slice(0, -2) + 'Id';
    } else if (result.endsWith('HP')) {
      result = result.slice(0, -2) + 'Hp';
    } else if (result.endsWith('AC')) {
      result = result.slice(0, -2) + 'Ac';
    }

    // Handle acronyms at the start of strings (e.g., HPValue -> hpValue)
    if (result.startsWith('HP')) {
      result = 'hp' + result.slice(2);
    } else if (result.startsWith('AC')) {
      result = 'ac' + result.slice(2);
    }

    // Lowercase the first character for standard camelCase
    return result.charAt(0).toLowerCase() + result.slice(1);
  }

  /**
   * Builds a hierarchical tree from a flat list of events.
   * Structure: Round -> Turn (SequenceID) -> Event -> Sub-events (ParentID)
   */
  private buildTree(events: SimulationEvent[]): TimelineNode[] {
    const roundsMap = new Map<number, TimelineNode>();
    const turnsMap = new Map<string, TimelineNode>();
    const nodesMap = new Map<string, TimelineNode>();

    // 1. Pre-create nodes for all actual events to facilitate hierarchy lookup
    events.forEach((event) => {
      nodesMap.set(event.id, { data: event, children: [] });
    });

    const rootNodes: TimelineNode[] = [];

    events.forEach((event) => {
      const node = nodesMap.get(event.id)!;

      // 2. Ensure Round node exists
      if (!roundsMap.has(event.round)) {
        const roundNode: TimelineNode = {
          data: {
            round: event.round,
            type: 'round',
            id: `round-${event.round}`,
            data: { note: `Round ${event.round}` },
          },
          children: [],
          expanded: event.round <= 1, // Auto-expand early rounds for better UX
        };
        roundsMap.set(event.round, roundNode);
        rootNodes.push(roundNode);
      }
      const roundNode = roundsMap.get(event.round)!;

      // 3. Handle Turns (grouped by sequenceId)
      if (event.sequenceId) {
        if (!turnsMap.has(event.sequenceId)) {
          const turnNode: TimelineNode = {
            data: {
              round: event.round,
              type: 'turn',
              id: event.sequenceId,
              sequenceId: event.sequenceId,
              data: {
                actor: event.data.actor,
                note: `Turn: ${event.data.actor?.name || 'Unknown'}`,
              },
            },
            children: [],
          };
          turnsMap.set(event.sequenceId, turnNode);
          roundNode.children?.push(turnNode);
        }
        const turnNode = turnsMap.get(event.sequenceId)!;

        // 4. Attach to hierarchy
        // If it has a parentId, and that parent is NOT the turn itself, nest it
        if (event.parentId && event.parentId !== event.sequenceId && nodesMap.has(event.parentId)) {
          nodesMap.get(event.parentId)!.children?.push(node);
        } else {
          // Top-level action within a turn
          turnNode.children?.push(node);
        }
      } else {
        // Events without a sequence (like Initiative) go directly under the round
        if (event.parentId && nodesMap.has(event.parentId)) {
          nodesMap.get(event.parentId)!.children?.push(node);
        } else {
          roundNode.children?.push(node);
        }
      }
    });

    return rootNodes;
  }
}
