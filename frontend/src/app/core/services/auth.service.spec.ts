import { TestBed } from '@angular/core/testing';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { AuthService } from './auth.service';
import { LoginResponse } from '../models/auth.model';

describe('AuthService (RF-11 IAM & RBAC)', () => {
  let service: AuthService;
  let httpMock: HttpTestingController;

  beforeEach(() => {
    localStorage.clear();
    TestBed.configureTestingModule({
      providers: [provideHttpClient(), provideHttpClientTesting(), provideRouter([])],
    });
    service = TestBed.inject(AuthService);
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
    httpMock.verify();
    localStorage.clear();
  });

  it('debe autenticar usuario, persistir token JWT y calcular ruta por rol', () => {
    const mockResp: LoginResponse = {
      accessToken: 'jwt.token.2026',
      tokenType: 'Bearer',
      expiresIn: 28800,
      usuario: {
        id: 12,
        nombreCompleto: 'Dr. Carlos Mendoza',
        email: 'medico@bloodlink.gob.pe',
        rol: 'MEDICO',
        hospitalId: 'HOSP-MINSA-01',
      },
    };

    service.login({ email: 'medico@bloodlink.gob.pe', password: 'Medico123!' }).subscribe((res) => {
      expect(res.accessToken).toBe('jwt.token.2026');
    });

    const req = httpMock.expectOne('/api/v1/auth/login');
    expect(req.request.method).toBe('POST');
    req.flush(mockResp);

    expect(service.isAuthenticated()).toBe(true);
    expect(service.currentRole()).toBe('MEDICO');
    expect(service.hasRole(['MEDICO'])).toBe(true);
    expect(service.hasRole(['PERSONAL_BANCO'])).toBe(false);
    expect(service.defaultRouteForRole('MEDICO')).toBe('/emergencias');
    expect(service.defaultRouteForRole('PERSONAL_BANCO')).toBe('/inventario');
    expect(service.defaultRouteForRole('COORDINADOR')).toBe('/dashboard');

    service.clearSession();
    expect(service.isAuthenticated()).toBe(false);
  });
});
