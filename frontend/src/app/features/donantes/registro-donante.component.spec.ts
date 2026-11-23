import { TestBed } from '@angular/core/testing';
import { Router, provideRouter } from '@angular/router';
import { of } from 'rxjs';
import { DonantesService } from '../../core/services/donantes.service';
import { RegistroDonanteComponent } from './registro-donante.component';

describe('RegistroDonanteComponent (CUN-002 / CP-03 / CP-05)', () => {
  let registrarLlamado = 0;

  const mockDonantesService = {
    validarDniReniec: (dni: string) =>
      of({
        dni,
        dniValido: true,
        nombreCompleto: 'JUAN CARLOS PEREZ GOMEZ',
        fechaNacimiento: '1995-04-12',
        mayorEdad: true,
      }),
    registrarDonante: () => {
      registrarLlamado++;
      return of({
        donanteId: 'DON-2026-A1B2C3D4',
        dni: '76543210',
        nombreCompleto: 'JUAN CARLOS PEREZ GOMEZ',
        tipoSangre: 'O' as const,
        factorRh: '+' as const,
        bancoPreferidoId: 'HOSP-MINSA-01',
        validadoReniec: true,
        habilitadoConvocatorias: true,
        fechaRegistro: '2026-11-22T20:00:00Z',
      });
    },
  };

  beforeEach(async () => {
    registrarLlamado = 0;
    await TestBed.configureTestingModule({
      imports: [RegistroDonanteComponent],
      providers: [
        provideRouter([]),
        { provide: DonantesService, useValue: mockDonantesService },
      ],
    }).compileComponents();
  });

  it('debe completar el flujo de 2 pasos autocompletando y bloqueando el nombre oficial (CP-03)', () => {
    const fixture = TestBed.createComponent(RegistroDonanteComponent);
    const comp = fixture.componentInstance as unknown as {
      dni: string;
      celular: string;
      correo: string;
      aceptaPoliticaPrivacidad: boolean;
      paso: () => string;
      validarIdentidad: () => void;
      completarRegistro: () => void;
      donanteRegistrado: () => { donanteId: string } | null;
    };

    fixture.detectChanges();
    expect(comp.paso()).toBe('IDENTIDAD');

    comp.dni = '76543210';
    comp.validarIdentidad();
    fixture.detectChanges();
    expect(comp.paso()).toBe('DATOS_CONTACTO');

    const nombreInput = fixture.nativeElement.querySelector('#nombreOficial') as HTMLInputElement;
    expect(nombreInput).toBeTruthy();
    expect(nombreInput.readOnly).toBe(true);
    expect(nombreInput.value).toBe('JUAN CARLOS PEREZ GOMEZ');

    comp.celular = '987654321';
    comp.correo = 'jperez@ciudadano.pe';
    comp.aceptaPoliticaPrivacidad = true;
    comp.completarRegistro();
    fixture.detectChanges();

    expect(registrarLlamado).toBe(1);
    expect(comp.paso()).toBe('CONFIRMADO');
    expect(comp.donanteRegistrado()?.donanteId).toBe('DON-2026-A1B2C3D4');
  });

  it('debe bloquear el envío y mostrar errores inline si celular o correo son inválidos', () => {
    const fixture = TestBed.createComponent(RegistroDonanteComponent);
    const comp = fixture.componentInstance as unknown as {
      dni: string;
      celular: string;
      correo: string;
      aceptaPoliticaPrivacidad: boolean;
      errorCelular: () => string;
      errorCorreo: () => string;
      errorPrivacidad: () => string;
      validarIdentidad: () => void;
      completarRegistro: () => void;
    };

    comp.dni = '76543210';
    comp.validarIdentidad();
    fixture.detectChanges();

    comp.celular = '81234567'; // No empieza con 9
    comp.correo = 'correo-invalido';
    comp.aceptaPoliticaPrivacidad = false;
    comp.completarRegistro();
    fixture.detectChanges();

    expect(registrarLlamado).toBe(0);
    expect(comp.errorCelular()).toBe('Ingresa un celular de 9 dígitos que empiece con 9.');
    expect(comp.errorCorreo()).toBe('Ingresa un correo electrónico válido.');
    expect(comp.errorPrivacidad()).toBe('Debes aceptar para continuar.');
  });

  it('debe mostrar el modal de confirmación al pulsar Volver con datos ingresados y navegar solo al confirmar', () => {
    const fixture = TestBed.createComponent(RegistroDonanteComponent);
    const router = TestBed.inject(Router);
    let navigatedUrl: string | null = null;
    router.navigateByUrl = (url) => {
      navigatedUrl = typeof url === 'string' ? url : url.toString();
      return Promise.resolve(true);
    };

    const comp = fixture.componentInstance as unknown as {
      dni: string;
    };

    comp.dni = '76543210';
    fixture.detectChanges();

    const backLink = fixture.nativeElement.querySelector('.back-link') as HTMLAnchorElement;
    backLink.click();
    fixture.detectChanges();

    expect(navigatedUrl).toBeNull();
    const dialog = fixture.nativeElement.querySelector('[role="dialog"]') as HTMLElement | null;
    expect(dialog).toBeTruthy();
    expect(dialog?.textContent).toContain('¿Salir sin guardar?');

    const stayBtn = fixture.nativeElement.querySelector('.modal-btn-stay') as HTMLButtonElement;
    stayBtn.click();
    fixture.detectChanges();

    expect(fixture.nativeElement.querySelector('[role="dialog"]')).toBeNull();
    expect(navigatedUrl).toBeNull();
    expect(comp.dni).toBe('76543210');

    backLink.click();
    fixture.detectChanges();
    const leaveBtn = fixture.nativeElement.querySelector('.modal-btn-leave') as HTMLButtonElement;
    leaveBtn.click();
    fixture.detectChanges();

    expect(navigatedUrl).toBe('/');
  });
});
