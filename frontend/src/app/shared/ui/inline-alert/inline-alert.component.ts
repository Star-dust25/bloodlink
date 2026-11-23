import { ChangeDetectionStrategy, Component, input } from '@angular/core';

export type InlineAlertVariant = 'info' | 'warning' | 'error' | 'success';

@Component({
  selector: 'app-inline-alert',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="inline-alert" [class]="'inline-alert variant-' + variant()" role="alert">
      <div class="alert-header">
        <span class="alert-dot" aria-hidden="true"></span>
        <span class="alert-title">{{ title() }}</span>
      </div>
      @if (message()) {
        <p class="alert-message">{{ message() }}</p>
      }
      <ng-content />
    </div>
  `,
  styles: [`
    .inline-alert {
      padding: 1.15rem 1.4rem;
      border: 1.5px solid var(--border-subtle);
      border-radius: var(--radius-subcard);
      background-color: var(--bg-surface);
      display: flex;
      flex-direction: column;
      gap: 0.35rem;
    }
    .variant-info {
      background-color: var(--brand-turquoise-soft);
      border-color: rgba(58, 161, 179, 0.35);
    }
    .variant-warning {
      background-color: var(--clinical-warning-bg);
      border-color: rgba(201, 122, 22, 0.35);
    }
    .variant-error {
      background-color: var(--clinical-critical-bg);
      border-color: rgba(201, 53, 63, 0.35);
    }
    .variant-success {
      background-color: var(--clinical-ok-bg);
      border-color: rgba(30, 142, 104, 0.35);
    }
    .alert-header {
      display: flex;
      align-items: center;
      gap: 0.6rem;
    }
    .alert-dot {
      width: 10px;
      height: 10px;
      border-radius: 50%;
      flex-shrink: 0;
      background-color: var(--brand-turquoise-dark);
    }
    .variant-warning .alert-dot { background-color: var(--state-warning); }
    .variant-error .alert-dot { background-color: var(--state-danger); }
    .variant-success .alert-dot { background-color: var(--state-ok); }
    .alert-title {
      font-size: 0.94rem;
      font-weight: 800;
      color: var(--text-primary);
    }
    .variant-error .alert-title { color: var(--state-danger); }
    .variant-warning .alert-title { color: var(--state-warning); }
    .variant-success .alert-title { color: var(--state-ok); }
    .alert-message {
      font-size: 0.9rem;
      font-weight: 500;
      color: var(--text-secondary);
      line-height: 1.55;
      padding-left: 1.15rem;
    }
  `],
})
export class InlineAlertComponent {
  readonly variant = input<InlineAlertVariant>('info');
  readonly title = input.required<string>();
  readonly message = input<string>('');
  readonly code = input<string>('');
}
