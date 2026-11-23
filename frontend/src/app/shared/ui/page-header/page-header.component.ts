import { ChangeDetectionStrategy, Component, input } from '@angular/core';

@Component({
  selector: 'app-page-header',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <header class="page-header">
      <div class="header-copy">
        @if (eyebrow()) {
          <span class="eyebrow">{{ eyebrow() }}</span>
        }
        <h1 class="page-title">{{ title() }}</h1>
        @if (subtitle()) {
          <p class="page-subtitle">{{ subtitle() }}</p>
        }
      </div>
      <div class="header-actions no-print">
        <ng-content />
      </div>
    </header>
  `,
  styles: [`
    .page-header {
      display: flex;
      flex-wrap: wrap;
      align-items: center;
      justify-content: space-between;
      gap: 1.25rem;
      margin-bottom: 1.5rem;
    }
    .header-copy {
      display: flex;
      flex-direction: column;
      gap: 0.35rem;
      max-width: 720px;
    }
    .page-title {
      font-size: clamp(1.65rem, 2.5vw, 2.05rem);
      font-weight: 900;
      letter-spacing: -0.025em;
      line-height: 1.12;
      color: var(--text-primary);
    }
    .page-subtitle {
      font-size: 0.96rem;
      font-weight: 500;
      color: var(--text-secondary);
      line-height: 1.55;
    }
    .header-actions {
      display: flex;
      align-items: center;
      gap: 0.75rem;
    }
  `],
})
export class PageHeaderComponent {
  readonly eyebrow = input<string>('');
  readonly title = input.required<string>();
  readonly subtitle = input<string>('');
}
