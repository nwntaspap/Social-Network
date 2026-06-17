'use client';

/**
 * app/register/page.tsx  →  route: /register
 *
 * Mirrors renderRegisterPage() + initFormBehaviour() from pages/register.js.
 *
 * On success → navigate to /login (same as the old JS).
 * On failure → map the backend error message to the relevant field.
 */

import Image from 'next/image';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { api, ApiError } from '@/lib/api';

type Gender = '' | 'male' | 'female' | 'other' | 'prefer_not_to_say';

interface RegisterBody {
  nickname: string;
  firstname: string;
  lastname: string;
  age: number;
  gender: Gender;
  email: string;
  password: string;
}

export default function RegisterPage() {
  const { user, loading } = useAuth();
  const router = useRouter();

  // Already logged in — bounce to home
  useEffect(() => {
    if (!loading && user) {
      router.replace('/');
    }
  }, [user, loading, router]);

  if (loading || user) {
    return (
      <div className="page-loading">
        <div className="page-loading-spinner" />
      </div>
    );
  }

  return <RegisterForm />;
}

// ─── Form ─────────────────────────────────────────────────────────────────────

function RegisterForm() {
  const router = useRouter();
  const [submitting, setSubmitting] = useState(false);
  const [showPassword, setShowPassword] = useState(false);

  // Field values
  const [nickname, setNickname] = useState('');
  const [firstname, setFirstname] = useState('');
  const [lastname, setLastname] = useState('');
  const [age, setAge] = useState('');
  const [gender, setGender] = useState<Gender>('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  // Field errors
  const [nicknameError, setNicknameError] = useState('');
  const [firstnameError, setFirstnameError] = useState('');
  const [lastnameError, setLastnameError] = useState('');
  const [ageError, setAgeError] = useState('');
  const [genderError, setGenderError] = useState('');
  const [emailError, setEmailError] = useState('');
  const [passwordError, setPasswordError] = useState('');

  function clearAllErrors() {
    setNicknameError('');
    setFirstnameError('');
    setLastnameError('');
    setAgeError('');
    setGenderError('');
    setEmailError('');
    setPasswordError('');
  }

  function handleReset() {
    setNickname('');
    setFirstname('');
    setLastname('');
    setAge('');
    setGender('');
    setEmail('');
    setPassword('');
    setShowPassword(false);
    clearAllErrors();
  }

  // Maps a backend error message to the correct field — mirrors mapErrorToField()
  function mapError(msg: string) {
    const lower = msg.toLowerCase();
    if (lower.includes('nickname') || lower.includes('nick')) {
      setNicknameError(msg);
    } else if (lower.includes('first')) {
      setFirstnameError(msg);
    } else if (lower.includes('last')) {
      setLastnameError(msg);
    } else if (lower.includes('age')) {
      setAgeError(msg);
    } else if (lower.includes('gender')) {
      setGenderError(msg);
    } else if (lower.includes('email')) {
      setEmailError(msg);
    } else {
      setPasswordError(msg);
    }
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    clearAllErrors();

    // Client-side validation — mirrors the old JS guard chain
    // Stop at first error so the user sees one problem at a time
    if (!nickname) {
      setNicknameError('Please enter a nickname.');
      return;
    }
    if (!firstname) {
      setFirstnameError('Please enter your first name.');
      return;
    }
    if (!lastname) {
      setLastnameError('Please enter your last name.');
      return;
    }
    if (!age) {
      setAgeError('Please enter your age.');
      return;
    }
    if (!gender) {
      setGenderError('Please select a gender.');
      return;
    }
    if (!email) {
      setEmailError('Please enter an email address.');
      return;
    }
    if (!password) {
      setPasswordError('Please enter a password.');
      return;
    }

    const body: RegisterBody = {
      nickname,
      firstname,
      lastname,
      age: Number(age),
      gender,
      email,
      password,
    };

    setSubmitting(true);
    try {
      await api.post('/register', body);
      // Success → go to login, same as the old JS
      router.push('/login');
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      mapError(msg);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <>
      <div className="page-title-box">
        <h1>Welcome to SocialNet</h1>
      </div>

      <div className="signup-container">
        <div className="signup-wrapper">
          <h2 className="signup-title">Sign Up</h2>
          <div className="text-base">
            Already a member? <Link href="/login">Sign In</Link>
          </div>

          {/* OAuth providers */}
          <div className="btn-box">
            <a className="signup-provider-btn google" href="/api/v1/auth/google/login">
              <Image src="/images/icons/google-logo.png" alt="Google Logo" width={20} height={20} />
              <p>Sign up with Google</p>
            </a>
            <a className="signup-provider-btn github" href="/api/v1/auth/github/login">
              <Image
                src="/images/icons/github-white-logo.png"
                alt="Github Logo"
                width={20}
                height={20}
              />
              <p>Sign up with Github</p>
            </a>
          </div>

          <div className="border">
            <span className="border-text">or</span>
          </div>

          <form className="signup" onSubmit={handleSubmit} noValidate>
            <div className="input-wrapper">
              <div className="input-box">
                <label htmlFor="nickname">Nickname</label>
                <input
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

              <div className="input-box">
                <label htmlFor="firstname">First Name</label>
                <input
                  type="text"
                  id="firstname"
                  name="firstname"
                  className="form-input"
                  placeholder="Enter your first name"
                  value={firstname}
                  onChange={(e) => {
                    setFirstname(e.target.value);
                    setFirstnameError('');
                  }}
                  autoComplete="given-name"
                />
                {firstnameError && <span className="error-message">{firstnameError}</span>}
              </div>

              <div className="input-box">
                <label htmlFor="lastname">Last Name</label>
                <input
                  type="text"
                  id="lastname"
                  name="lastname"
                  className="form-input"
                  placeholder="Enter your last name"
                  value={lastname}
                  onChange={(e) => {
                    setLastname(e.target.value);
                    setLastnameError('');
                  }}
                  autoComplete="family-name"
                />
                {lastnameError && <span className="error-message">{lastnameError}</span>}
              </div>

              <div className="input-box">
                <label htmlFor="age">Age</label>
                <input
                  type="number"
                  id="age"
                  name="age"
                  min={0}
                  className="form-input"
                  placeholder="Enter your age"
                  value={age}
                  onChange={(e) => {
                    setAge(e.target.value);
                    setAgeError('');
                  }}
                />
                {ageError && <span className="error-message">{ageError}</span>}
              </div>

              <div className="input-box">
                <label htmlFor="gender">Gender</label>
                <select
                  id="gender"
                  name="gender"
                  className="form-input form-select"
                  value={gender}
                  onChange={(e) => {
                    setGender(e.target.value as Gender);
                    setGenderError('');
                  }}
                >
                  <option value="">Select gender</option>
                  <option value="male">Male</option>
                  <option value="female">Female</option>
                  <option value="other">Other</option>
                  <option value="prefer_not_to_say">Prefer not to say</option>
                </select>
                {genderError && <span className="error-message">{genderError}</span>}
              </div>

              <div className="input-box">
                <label htmlFor="email">E-mail</label>
                <input
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
                    autoComplete="new-password"
                  />
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
                {submitting ? 'Signing up…' : 'Sign Up'}
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
