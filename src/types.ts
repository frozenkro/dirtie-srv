import 'dotenv/config';
import 'pino';
import { EventEmitter } from 'events';

export interface LoadTestEvent extends EventEmitter {
  on(event: string, listener: (...args: any[]) => void): this;
  emit(event: string, ...args: any[]): boolean;
}

export interface LoadTestEventData {
  timestamp: number;
  type: 'success' | 'error';
  latency?: number;
  error?: string;
  payload?: any;
}
