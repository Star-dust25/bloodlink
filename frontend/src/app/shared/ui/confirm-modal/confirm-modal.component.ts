import { ChangeDetectionStrategy, Component, input, output } from '@angular/core';

@Component({
  selector: 'app-confirm-modal',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  host: {
    '(document:keydown.escape)': 'onEscape($event)',
  },
  template: `
    @if (open()) {
      <div class="modal-backdrop" (click)="onBackdropClick()">
        <div
          class="modal-card"
          role="dialog"
          aria-modal="true"
          aria-labelledby="confirmModalTitle"
          aria-describedby="confirmModalDesc"
          (click)="$event.stopPropagation()"
        >
          <h2 id="confirmModalTitle" class="modal-title">{{ title() }}</h2>
          <p id="confirmModalDesc" class="modal-message">{{ message() }}</p>

          <div class="modal-actions">
            <button
              type="button"
              class="modal-btn modal-btn-stay"
              (click)="cancel.emit()"
            >
              {{ cancelLabel() }}
            </button>
            <button
              type="button"
              class="modal-btn modal-btn-leave"
              (click)="confirm.emit()"
            >
              {{ confirmLabel() }}
            </button>
          </div>
        </div>
      </div>
    }
  `,
  styles: [`
    .modal-backdrop {
      position: fixed;
      inset: 0;
      z-index: 1000;
      background: rgba(34, 30, 30, 0.45);
      backdrop-filter: blur(4px);
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 1rem;
      animation: modalBackdropFade 200ms ease forwards;
    }

    .modal-card {
      max-width: 440px;
      width: calc(100% - 2rem);
      border-radius: 28px;
      background: #FFFFFF;
      padding: 2rem;
      box-shadow: 0 24px 48px rgba(34, 30, 30, 0.16);
      display: flex;
      flex-direction: column;
      gap: 0.85rem;
      animation: modalCardSlide 220ms cubic-bezier(0.22, 1, 0.36, 1) forwards;
    }

    .modal-title {
      font-family: var(--font-sans);
      font-weight: 800;
      font-size: 1.35rem;
      line-height: 1.25;
      color: #221E1E;
      margin: 0;
    }

    .modal-message {
      font-size: 1rem;
      color: #534D4A;
      line-height: 1.55;
      margin: 0;
    }

    .modal-actions {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      margin-top: 0.65rem;
    }

    .modal-btn {
      flex: 1;
      min-height: 50px;
      padding: 0 1.25rem;
      border-radius: 9999px;
      border: none;
      font-family: var(--font-sans);
      font-weight: 800;
      font-size: 1rem;
      cursor: pointer;
      transition: background-color 180ms ease, transform 160ms ease;
    }

    .modal-btn-stay {
      background-color: #F7CC51;
      color: #221E1E;
    }

    .modal-btn-stay:hover {
      background-color: var(--brand-yellow-hover);
    }

    .modal-btn-leave {
      background-color: #F3EFE4;
      color: #221E1E;
    }

    .modal-btn-leave:hover {
      background-color: #EAE4D5;
    }

    .modal-btn:focus-visible {
      outline: 3px solid var(--brand-turquoise);
      outline-offset: 2px;
    }

    @keyframes modalBackdropFade {
      from { opacity: 0; }
      to { opacity: 1; }
    }

    @keyframes modalCardSlide {
      from { opacity: 0; transform: translateY(10px); }
      to { opacity: 1; transform: translateY(0); }
    }

    @media (max-width: 480px) {
      .modal-card {
        padding: 1.65rem;
      }
      .modal-actions {
        flex-direction: column;
        align-items: stretch;
      }
    }
  `],
})
export class ConfirmModalComponent {
  readonly open = input<boolean>(false);
  readonly title = input<string>('¿Salir sin guardar?');
  readonly message = input<string>('Si vuelves al inicio ahora, perderás los datos que ingresaste.');
  readonly confirmLabel = input<string>('Sí, salir');
  readonly cancelLabel = input<string>('Seguir aquí');

  readonly confirm = output<void>();
  readonly cancel = output<void>();

  onBackdropClick(): void {
    this.cancel.emit();
  }

  onEscape(event: Event): void {
    if (!this.open()) return;
    event.stopPropagation();
    this.cancel.emit();
  }
}
