import { ChangeDetectionStrategy, Component, input } from '@angular/core';

@Component({
  selector: 'app-empty-state',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `
    <div class="empty-state">
      <span class="eyebrow">{{ eyebrow() }}</span>
      <h3 class="empty-title">{{ title() }}</h3>
      @if (description()) {
        <p class="empty-desc">{{ description() }}</p>
      }
      <div class="empty-actions">
        <ng-content />
      </div>
    </div>
  `,
  styles: [`
    .empty-state {
      padding: 3rem 2rem;
      background-color: var(--bg-surface);
      border: 2px dashed var(--brand-sand);
      border-radius: var(--radius-card);
      text-align: center;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 0.55rem;
    }
    .empty-title {
      font-size: 1.12rem;
      font-weight: 800;
      color: var(--text-primary);
    }
    .empty-desc {
      font-size: 0.92rem;
      font-weight: 500;
      color: var(--text-muted);
      max-width: 460px;
    }
    .empty-actions {
      margin-top: 0.85rem;
    }
  `],
})
export class EmptyStateComponent {
  readonly eyebrow = input<string>('SIN REGISTROS');
  readonly title = input.required<string>();
  readonly description = input<string>('');
}
