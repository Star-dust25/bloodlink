export type RolUsuario = 'MEDICO' | 'PERSONAL_BANCO' | 'COORDINADOR';

export interface LoginRequest {
  email: string;
  password: string;
}

export interface UsuarioDTO {
  id: number;
  nombreCompleto: string;
  email: string;
  rol: RolUsuario;
  hospitalId: string;
}

export interface LoginResponse {
  accessToken: string;
  tokenType: string;
  expiresIn: number;
  usuario: UsuarioDTO;
}
