import { Action, Actor, ActorSummary, MultiattackOption } from './actor.model';

export function isCharacter(actor: Actor | null | undefined): boolean {
  if (!actor) return false;
  // Check new Actor-based Character
  if (actor.actorType === 'character') return true;
  if (actor.metadata?.classId !== undefined) return true;
  return false;
}

export function isMonster(actor: Actor | null | undefined): boolean {
  if (!actor) return false;
  // Check new Actor-based Monster
  if (actor.actorType === 'monster') return true;
  if (actor.metadata?.cr !== undefined && actor.metadata?.classId === undefined) return true;
  return false;
}

export * from '../core';
export * from './actor.model';
