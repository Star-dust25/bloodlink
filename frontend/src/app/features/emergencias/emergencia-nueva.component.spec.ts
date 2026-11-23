import { HttpErrorResponse } from '@angular/common/http';
import { TestBed } from '@angular/core/testing';
import { provideRouter } from '@angular/router';
import { throwError } from 'rxjs';
import { AuthService } from '../../core/services/auth.service';
import { EmergenciasService } from '../../core/services/emergencias.service';
import { EmergenciaNuevaComponent } from './emergencia-nueva.component';

describe('EmergenciaNuevaComponent (CUN-001 / CP-02 / 409 Conflict)', () => {
  const mockEmergenciasService = {
    crearSolicitudYMatching: () =>
      throwError(
        () =>
          new HttpErrorResponse({
            status: 409,
            error: {
              solicitudId: 'SOL-2026-CONFLICT1',
              estadoSolicitud: 'En Espera de Donantes',
              mensaje: 'Stock no disponible en red',
              sugerirConvocatoriaDonantes: true,
              tiempoRespuestaMatchingMs: 18,
              fechaHora: '2026-11-22T20:00:00Z',
            },
          })
      ),
  };

  const mockAuthService = {
    currentUser: () => ({
      id: 12,
      nombreCompleto: 'Dr. Carlos Mendoza',
      email: 'medico@bloodlink.gob.pe',
      rol: 'MEDICO',
      hospitalId: 'HOSP-MINSA-01',
    }),
  };

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EmergenciaNuevaComponent],
      providers: [
        provideRouter([]),
        { provide: EmergenciasService, useValue: mockEmergenciasService },
        { provide: AuthService, useValue: mockAuthService },
      ],
    }).compileComponents();
  });

  it('debe capturar respuesta 409 Conflict y activar vista de convocatoria de donantes', () => {
    const fixture = TestBed.createComponent(EmergenciaNuevaComponent);
    const comp = fixture.componentInstance as unknown as {
      ejecutarMatching: () => void;
      conflictoSinStock: () => { solicitudId: string; sugerirConvocatoriaDonantes: boolean } | null;
    };

    comp.ejecutarMatching();
    expect(comp.conflictoSinStock()?.solicitudId).toBe('SOL-2026-CONFLICT1');
    expect(comp.conflictoSinStock()?.sugerirConvocatoriaDonantes).toBe(true);
  });
});
