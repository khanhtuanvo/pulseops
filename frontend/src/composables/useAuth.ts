import router from '@/router';
import { useAuthStore } from '@/stores/auth';

const verifierKey = 'pkce_verifier';
const stateKey = 'oauth_state';
const apiBaseUrl = () => import.meta.env.VITE_API_URL ?? '';

function base64Url(bytes: ArrayBuffer | Uint8Array) {
  const array = bytes instanceof Uint8Array ? bytes : new Uint8Array(bytes);
  let binary = '';
  array.forEach((byte) => {
    binary += String.fromCharCode(byte);
  });
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function randomHex(byteCount: number) {
  const bytes = new Uint8Array(byteCount);
  window.crypto.getRandomValues(bytes);
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
}

export async function createPkcePair() {
  const verifierBytes = new Uint8Array(43);
  window.crypto.getRandomValues(verifierBytes);
  const codeVerifier = base64Url(verifierBytes);
  const challengeInput = new TextEncoder().encode(codeVerifier);
  const challengeHash = await window.crypto.subtle.digest('SHA-256', challengeInput);
  return {
    codeVerifier,
    codeChallenge: base64Url(challengeHash),
  };
}

export function useAuth() {
  const authStore = useAuthStore();

  async function login() {
    const { codeVerifier, codeChallenge } = await createPkcePair();
    const state = randomHex(16);
    sessionStorage.setItem(verifierKey, codeVerifier);
    sessionStorage.setItem(stateKey, state);

    const redirectUri = new URL('/auth/callback', window.location.origin).toString();
    const url = new URL('https://accounts.google.com/o/oauth2/v2/auth');
    url.searchParams.set('client_id', import.meta.env.VITE_GOOGLE_CLIENT_ID);
    url.searchParams.set('redirect_uri', redirectUri);
    url.searchParams.set('response_type', 'code');
    url.searchParams.set('scope', 'openid email profile');
    url.searchParams.set('state', state);
    url.searchParams.set('code_challenge', codeChallenge);
    url.searchParams.set('code_challenge_method', 'S256');
    window.location.href = url.toString();
  }

  async function handleCallback(code: string, state: string) {
    const expectedState = sessionStorage.getItem(stateKey);
    if (!expectedState || expectedState !== state) {
      throw new Error('OAuth state mismatch');
    }

    const codeVerifier = sessionStorage.getItem(verifierKey);
    sessionStorage.removeItem(verifierKey);
    sessionStorage.removeItem(stateKey);
    if (!codeVerifier) {
      throw new Error('Missing PKCE verifier');
    }

    const response = await fetch(`${apiBaseUrl()}/auth/callback`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code, codeVerifier, state }),
    });
    if (!response.ok) {
      throw new Error('Authentication failed');
    }

    await authStore.fetchMe();
    const redirect = router.currentRoute.value.query.redirect;
    await router.push(typeof redirect === 'string' ? redirect : '/dashboard');
  }

  return { login, handleCallback };
}
