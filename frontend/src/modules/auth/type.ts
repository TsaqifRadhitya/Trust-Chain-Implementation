export interface User {
    email: string;
    name: string;
}

export interface LoginResponse {
    success: boolean;
    user?: User | null;
    error?: string;
}
