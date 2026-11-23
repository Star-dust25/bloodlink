import { ChangeDetectionStrategy, Component, input } from '@angular/core';

export type IndicatorTone = 'neutral' | 'ok' | 'warning' | 'critical' | 'ink';

@Component({
  selector: 'app-status-indicator',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <span class="status-indicator">
      <span class="dot" [class]="'dot-' + tone()" aria-hidden="true"></span>
      <span class="label">{{ label() }}</span>
      @if (detail()) {
        <span class="detail tabular">({{ detail() }})</span>
      }
    </span>
  `,
  styles: [`
    .status-indicator {
      display: inline-flex;
      align-items: center;
      gap: 0.55rem;
      font-size: 0.88rem;
      font-weight: 700;
      color: var(--text-primary);
      white-space: nowrap;
    }
    .dot {
      width: 10px;
      height: 10px;
      border-radius: 50%;
      flex-shrink: 0;
      background-color: var(--brand-sand);
    }
    .dot-ink {
      background-color: var(--brand-turquoise-dark);
      box-shadow: 0 0 0 3px var(--brand-turquoise-soft);
    }
    .dot-ok {
      background-color: var(--state-ok);
      box-shadow: 0 0 0 3px var(--clinical-ok-bg);
    }
    .dot-warning {
      background-color: var(--state-warning);
      box-shadow: 0 0 0 3px var(--clinical-warning-bg);
    }
    .dot-critical {
      background-color: var(--state-danger);
      box-shadow: 0 0 0 3px var(--clinical-critical-bg);
    }
    .detail {
      font-size: 0.8rem;
      font-weight: 600;
      color: var(--text-muted);
    }
  `],
})
export class StatusIndicatorComponent {
  readonly tone = input<IndicatorTone>('neutral');
  readonly label = input.required<string>();
  readonly detail = input<string>('');
}
