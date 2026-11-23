import { ChangeDetectionStrategy, Component, computed, input, output, signal } from '@angular/core';

export interface SelectOption {
  value: string;
  label: string;
  subtitle?: string;
}

@Component({
  selector: 'app-custom-select',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  host: {
    '(document:click)': 'close()',
    '(document:keydown.escape)': 'close()',
  },
  template: `
    <div class="custom-select" [class.open]="isOpen()">
      <button
        type="button"
        class="select-trigger"
        role="combobox"
        aria-haspopup="listbox"
        [attr.aria-label]="ariaLabel()"
        [attr.aria-expanded]="isOpen()"
        (click)="toggle($event)"
        (keydown)="onKeyDown($event)"
      >
        <span class="trigger-copy">
          @if (selectedOption(); as opt) {
            <span class="trigger-label">{{ opt.label }}</span>
            @if (opt.subtitle) {
              <span class="trigger-sub">{{ opt.subtitle }}</span>
            }
          } @else {
            <span class="trigger-placeholder">{{ placeholder() }}</span>
          }
        </span>

        <span class="trigger-chevron" aria-hidden="true">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
            <path d="M6 9L12 15L18 9" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" />
          </svg>
        </span>
      </button>

      <div
        class="select-menu"
        [class.placement-top]="placement() === 'top'"
        role="listbox"
        [attr.aria-label]="ariaLabel()"
      >
        @for (opt of options(); track opt.value; let i = $index) {
          <button
            type="button"
            class="select-option"
            role="option"
            [class.selected]="opt.value === value()"
            [class.highlighted]="i === highlightedIndex()"
            [attr.aria-selected]="opt.value === value()"
            (click)="selectOption(opt, $event)"
          >
            <span class="opt-copy">
              <span class="opt-label">{{ opt.label }}</span>
              @if (opt.subtitle) {
                <span class="opt-sub">{{ opt.subtitle }}</span>
              }
            </span>

            @if (opt.value === value()) {
              <span class="opt-check" aria-hidden="true">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none">
                  <path d="M5 13L9.5 17.5L19 7" stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
              </span>
            }
          </button>
        }
      </div>
    </div>
  `,
  styles: [`
    .custom-select { position: relative; width: 100%; }
    .select-trigger {
      width: 100%; min-height: 54px; padding: 12px 18px;
      border-radius: 16px; border: 1.5px solid var(--border-subtle);
      background-color: var(--bg-surface-muted); color: var(--text-primary);
      display: flex; align-items: center; justify-content: space-between; gap: 0.75rem;
      cursor: pointer; text-align: left; transition: all 0.2s ease;
    }
    .custom-select.open .select-trigger,
    .select-trigger:focus-visible {
      outline: none; background-color: #FFFFFF;
      border-color: var(--brand-turquoise);
      box-shadow: 0 0 0 4px rgba(58, 161, 179, 0.15);
    }
    .trigger-copy { display: flex; flex-direction: column; line-height: 1.3; min-width: 0; }
    .trigger-label { font-size: 0.98rem; font-weight: 700; color: var(--text-primary); }
    .trigger-sub { font-size: 0.8rem; font-weight: 600; color: var(--text-muted); }
    .trigger-placeholder { font-size: 0.98rem; font-weight: 600; color: var(--text-muted); }
    .trigger-chevron {
      color: var(--text-secondary); display: inline-flex; flex-shrink: 0;
      transition: transform 0.25s ease;
    }
    .custom-select.open .trigger-chevron {
      transform: rotate(180deg); color: var(--brand-turquoise-dark);
    }
    .select-menu {
      position: absolute; left: 0; right: 0; top: calc(100% + 8px); z-index: 40;
      background-color: #FFFFFF; border: 1px solid var(--border-subtle);
      border-radius: 20px; padding: 6px;
      box-shadow: 0 16px 36px rgba(34, 30, 30, 0.12);
      display: flex; flex-direction: column; gap: 2px;
      opacity: 0; visibility: hidden; pointer-events: none; transform: translateY(-4px);
      transition: opacity 200ms ease, transform 200ms ease, visibility 200ms ease;
    }
    .select-menu.placement-top {
      top: auto;
      bottom: calc(100% + 8px);
      transform: translateY(4px);
    }
    .custom-select.open .select-menu,
    .custom-select.open .select-menu.placement-top {
      opacity: 1; visibility: visible; pointer-events: auto; transform: translateY(0);
    }
    .select-option {
      width: 100%; padding: 11px 14px; border: none; border-radius: 14px;
      background: transparent; display: flex; align-items: center; justify-content: space-between;
      gap: 0.75rem; text-align: left; cursor: pointer; transition: background-color 150ms ease;
    }
    .select-option:hover, .select-option.highlighted { background-color: var(--bg-canvas); }
    .select-option.selected { background-color: var(--brand-turquoise-soft); color: var(--brand-turquoise-dark); }
    .opt-copy { display: flex; flex-direction: column; line-height: 1.3; }
    .opt-label { font-size: 0.96rem; font-weight: 700; color: var(--text-primary); }
    .select-option.selected .opt-label { color: var(--brand-turquoise-dark); }
    .opt-sub { font-size: 0.8rem; font-weight: 600; color: var(--text-muted); }
    .opt-check { color: var(--brand-turquoise-dark); display: inline-flex; flex-shrink: 0; }
  `],
})
export class CustomSelectComponent {
  readonly options = input.required<readonly SelectOption[]>();
  readonly value = input<string>('');
  readonly placeholder = input<string>('Selecciona una opción');
  readonly ariaLabel = input<string>('Selector');
  readonly placement = input<'bottom' | 'top'>('bottom');
  readonly valueChange = output<string>();

  protected readonly isOpen = signal(false);
  protected readonly highlightedIndex = signal(0);

  protected readonly selectedOption = computed(
    () => this.options().find((o) => o.value === this.value()) ?? null
  );

  toggle(event: Event): void {
    event.stopPropagation();
    const next = !this.isOpen();
    if (next) {
      const idx = this.options().findIndex((o) => o.value === this.value());
      this.highlightedIndex.set(idx >= 0 ? idx : 0);
    }
    this.isOpen.set(next);
  }

  close(): void {
    if (this.isOpen()) {
      this.isOpen.set(false);
    }
  }

  selectOption(opt: SelectOption, event?: Event): void {
    event?.stopPropagation();
    this.valueChange.emit(opt.value);
    this.isOpen.set(false);
  }

  onKeyDown(event: KeyboardEvent): void {
    const opts = this.options();
    if (opts.length === 0) return;

    switch (event.key) {
      case 'ArrowDown':
        event.preventDefault();
        if (!this.isOpen()) {
          this.isOpen.set(true);
        } else {
          this.highlightedIndex.update((i) => (i + 1) % opts.length);
        }
        break;
      case 'ArrowUp':
        event.preventDefault();
        if (!this.isOpen()) {
          this.isOpen.set(true);
        } else {
          this.highlightedIndex.update((i) => (i - 1 + opts.length) % opts.length);
        }
        break;
      case 'Enter':
      case ' ':
        event.preventDefault();
        if (!this.isOpen()) {
          this.isOpen.set(true);
        } else {
          const chosen = opts[this.highlightedIndex()];
          if (chosen) this.selectOption(chosen);
        }
        break;
      case 'Escape':
        this.close();
        break;
    }
  }
}
