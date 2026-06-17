'use client';

/**
 * app/login/page.tsx  →  route: /login
 *
 * Mirrors renderLoginPage() + initLoginBehaviour() from pages/login.js.
 *
 * Supports login by username (nickname) or email.
 * On success → navigate to / (AuthGuard there will confirm the session).
 * On failure → error messages shown under the relevant field.
 */

import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useEffect, useRef, useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { api, ApiError } from '@/lib/api';
import { User } from '@/lib/types';

type LoginType = 'username' | 'email';

export default function LoginPage() {
  const { user, loading, setUser } = useAuth();
  const router = useRouter();

  // If already logged in, go straight to home — mirrors the old BFF redirect
  useEffect(() => {
    if (!loading && user) {
      router.replace('/');
    }
  }, [user, loading, router]);

  // Don't flash the form while auth is resolving
  if (loading || user) {
    return (
      <div className="page-loading">
        <div className="page-loading-spinner" />
      </div>
    );
  }

  return <LoginForm setUser={setUser} />;
}

// ─── Form ─────────────────────────────────────────────────────────────────────

function LoginForm({ setUser }: { setUser: (u: User) => void }) {
  const router = useRouter();

  const [loginType, setLoginType] = useState<LoginType>('username');
  const [submitting, setSubmitting] = useState(false);

  // Field values
  const [nickname, setNickname] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  // Field errors
  const [nicknameError, setNicknameError] = useState('');
  const [emailError, setEmailError] = useState('');
  const [passwordError, setPasswordError] = useState('');

  // Focus the active input when loginType switches — mirrors updateVisibility()
  const nicknameRef = useRef<HTMLInputElement>(null);
  const emailRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (loginType === 'username') {
      setEmail('');
      setEmailError('');
      nicknameRef.current?.focus();
    } else {
      setNickname('');
      setNicknameError('');
      emailRef.current?.focus();
    }
  }, [loginType]);

  function handleReset() {
    setNickname('');
    setEmail('');
    setPassword('');
    setShowPassword(false);
    setNicknameError('');
    setEmailError('');
    setPasswordError('');
    setLoginType('username');
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    setNicknameError('');
    setEmailError('');
    setPasswordError('');

    // Client-side validation — mirrors the old JS guard block
    if (!password) {
      setPasswordError('Please enter your password.');
      return;
    }
    if (loginType === 'email' && !email) {
      setEmailError('Please enter your email address.');
      return;
    }
    if (loginType === 'username' && !nickname) {
      setNicknameError('Please enter your nickname.');
      return;
    }

    setSubmitting(true);
    try {
      let me: User;
      if (loginType === 'email') {
        me = await api.post<User>('/login/email', { email, password });
      } else {
        // Backend reads field "username" — mirrors the old api.post call
        me = await api.post<User>('/login/username', { username: nickname, password });
      }

      // Update auth context so Navbar re-renders immediately
      setUser(me);
      router.push('/');
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      const lower = msg.toLowerCase();

      if (lower.includes('email')) {
        setEmailError(msg);
      } else if (
        lower.includes('nickname') ||
        lower.includes('nick') ||
        lower.includes('username')
      ) {
        setNicknameError(msg);
      } else {
        setPasswordError(msg);
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      <div className="page-title-box">
        <h1>Welcome Back</h1>
      </div>

      <div className="signup-container">
        <div className="signup-wrapper">
          <h2 className="signup-title">Sign In</h2>
          <div className="text-base">
            Don&apos;t have an account? <Link href="/register">Sign Up</Link>
          </div>

          {/* OAuth providers */}
          <div className="btn-box">
            <a className="signup-provider-btn google" href="/api/v1/auth/google/login">
              <Image src="/images/icons/google-logo.png" alt="Google Logo" width={20} height={20} />
              <p>Continue with Google</p>
            </a>
            <a className="signup-provider-btn github" href="/api/v1/auth/github/login">
              <Image
                src="/images/icons/github-white-logo.png"
                alt="Github Logo"
                width={20}
                height={20}
              />
              <p>Continue with Github</p>
            </a>
          </div>

          <div className="border">
            <span className="border-text">or</span>
          </div>

          <form className="signup" onSubmit={handleSubmit} noValidate>
            <div className="input-wrapper">
              {/* Login type selector — mirrors .login-type-selector */}
              <div className="login-type-selector">
                <p className="login-type-title">Choose Type of Login</p>
                <label className="login-radio-label">
                  <input
                    type="radio"
                    name="loginType"
                    value="username"
                    className="login-type-radio"
                    checked={loginType === 'username'}
                    onChange={() => setLoginType('username')}
                  />
                  <span>Username</span>
                </label>
                <label className="login-radio-label">
                  <input
                    type="radio"
                    name="loginType"
                    value="email"
                    className="login-type-radio"
                    checked={loginType === 'email'}
                    onChange={() => setLoginType('email')}
                  />
                  <span>Email</span>
                </label>
              </div>

              {/* Nickname field — hidden when email mode */}
              <div
                className="input-box"
                style={{ display: loginType === 'username' ? undefined : 'none' }}
              >
                <label htmlFor="nickname">Nickname</label>
                <input
                  ref={nicknameRef}
                  type="text"
                  id="nickname"
                  name="nickname"
                  className="form-input"
                  placeholder="Enter your nickname"
                  value={nickname}
                  onChange={(e) => {
                    setNickname(e.target.value);
                    setNicknameError('');
                  }}
                  autoComplete="username"
                />
                {nicknameError && <span className="error-message">{nicknameError}</span>}
              </div>

              {/* Email field — hidden when username mode */}
              <div
                className="input-box"
                style={{ display: loginType === 'email' ? undefined : 'none' }}
              >
                <label htmlFor="email">Email address</label>
                <input
                  ref={emailRef}
                  type="email"
                  id="email"
                  name="email"
                  className="form-input"
                  placeholder="Enter your email"
                  value={email}
                  onChange={(e) => {
                    setEmail(e.target.value);
                    setEmailError('');
                  }}
                  autoComplete="email"
                />
                {emailError && <span className="error-message">{emailError}</span>}
              </div>

              {/* Password */}
              <div className="input-box">
                <div className="password-wrapper">
                  <label htmlFor="password">Password</label>
                  <input
                    type={showPassword ? 'text' : 'password'}
                    id="password"
                    name="password"
                    className="form-input"
                    placeholder="Enter your password"
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      setPasswordError('');
                    }}
                    autoComplete="current-password"
                  />
                  {/* Toggle visibility — mirrors the checkbox trick in the original */}
                  <button
                    type="button"
                    className="togglePassword"
                    aria-label={showPassword ? 'Hide password' : 'Show password'}
                    onClick={() => setShowPassword((prev) => !prev)}
                    style={{ background: 'none', border: 'none', cursor: 'pointer', padding: 0 }}
                  >
                    <Image
                      src={showPassword ? '/images/icons/hidden.png' : '/images/icons/eye.png'}
                      alt={showPassword ? 'Hide password' : 'Show password'}
                      className={showPassword ? 'hidden-icon' : 'eye-icon'}
                      width={20}
                      height={20}
                    />
                  </button>
                </div>
                {passwordError && <span className="error-message">{passwordError}</span>}
              </div>
            </div>

            <div className="btn-box">
              <button type="button" className="btn-reset-form" onClick={handleReset}>
                Reset Form
              </button>
              <button type="submit" className="btn-signup" disabled={submitting}>
                {submitting ? 'Signing in…' : 'Sign In'}
              </button>
            </div>
          </form>
        </div>

        <div className="home-link-container">
          <Link href="/" className="home-link">
            Go to Homepage
          </Link>
        </div>
      </div>
    </>
  );
}
