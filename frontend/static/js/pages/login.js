/**
 * pages/login.js
 *
 * SPA version of frontend/html/pages/login.html.
 * Supports login by nickname (username) or email. Sends POST to
 * /api/v1/login/username or /api/v1/login/email using the shared api client.
 */

import { api } from '../api.js';
import { navigate } from '../router.js';
import { escapeHTML } from '../helpers.js';

export async function renderLoginPage(user) {
  const root = document.getElementById('app-root');
  if (!root) return;

  root.innerHTML = buildLoginHTML();

  initLoginBehaviour();
}

function buildLoginHTML() {
  return /* html */ `
    <div class="page-title-box">
      <h1>Welcome Back</h1>
    </div>
    <div class="signup-container">
      <div class="signup-wrapper">
        <h2 class="signup-title">Sign In</h2>
        <div class="text-base">
          Don't have an account?
          <a href="/register" data-link>Sign Up</a>
        </div>
 
        <div class="btn-box">
          <a class="signup-provider-btn google" href="/auth/google/login" data-link>
            <img src="/static/images/icons/google-logo.png" alt="Google Logo" />
            <p>Continue with Google</p>
          </a>
          <a class="signup-provider-btn github" href="/auth/github/login" data-link>
            <img src="/static/images/icons/github-white-logo.png" alt="Github Logo" />
            <p>Continue with Github</p>
          </a>
        </div>
 
        <div class="border">
          <span class="border-text">or</span>
        </div>
 
        <form class="signup">
          <div class="input-wrapper">
 
            <div class="login-type-selector">
              <p class="login-type-title">Choose Type of Login</p>
              <label class="login-radio-label">
                <input type="radio" name="loginType" id="loginTypeUsername"
                       value="username" class="login-type-radio" checked />
                <span>Username</span>
              </label>
              <label class="login-radio-label">
                <input type="radio" name="loginType" id="loginTypeEmail"
                       value="email" class="login-type-radio" />
                <span>Email</span>
              </label>
            </div>
 
            <div class="input-box" id="nicknameBox">
              <label for="nickname">Nickname</label>
              <input type="text" name="nickname" id="nickname"
                     class="form-input" placeholder="Enter your nickname" />
              <span class="error-message" id="nickname-error"></span>
            </div>
 
            <div class="input-box" id="emailBox" style="display:none">
              <label for="email">Email address</label>
              <input type="email" name="email" id="email"
                     class="form-input" placeholder="Enter your email" />
              <span class="error-message" id="email-error"></span>
            </div>
 
            <div class="input-box">
              <div class="password-wrapper">
                <label for="password">Password</label>
                <input type="password" name="password" id="password"
                       class="form-input" placeholder="Enter your password" />
                <label for="togglePassword" class="togglePassword">
                  <img src="/static/images/icons/eye.png"    alt="Show password"
                       id="eye-icon"    class="eye-icon" />
                  <img src="/static/images/icons/hidden.png" alt="Hide password"
                       id="hidden-icon" class="hidden-icon" style="display:none" />
                </label>
                <input type="checkbox" id="togglePassword"
                       class="hidden-checkbox" aria-label="show/hide password" />
              </div>
              <span class="error-message" id="password-error"></span>
            </div>
 
          </div>
 
          <div class="btn-box">
            <button type="reset" class="btn-reset-form">Reset Form</button>
            <button type="submit" class="btn-signup">Sign In</button>
          </div>
        </form>
      </div>
 
      <div class="home-link-container">
        <a href="/" class="home-link" data-link>Go to Homepage</a>
      </div>
    </div>
  `;
}

function initLoginBehaviour() {
  const form = document.querySelector('.signup');
  if (!form) return;

  const loginTypeUsername = document.getElementById('loginTypeUsername');
  const loginTypeEmail = document.getElementById('loginTypeEmail');
  const nicknameBox = document.getElementById('nicknameBox');
  const emailBox = document.getElementById('emailBox');

  const nicknameInput = document.getElementById('nickname');
  const emailInput = document.getElementById('email');
  const passwordInput = document.getElementById('password');

  const nicknameError = document.getElementById('nickname-error');
  const emailError = document.getElementById('email-error');
  const passwordError = document.getElementById('password-error');

  // ---Login type toggle---
  function updateVisibility() {
    const useEmail = loginTypeEmail.checked;

    if (useEmail) {
      nicknameBox.style.display = 'none';
      emailBox.style.display = 'block';
      nicknameError.textContent = '';
      nicknameInput.value = '';
    } else {
      nicknameBox.style.display = '';
      emailBox.style.display = 'none';
      emailError.textContent = '';
      emailInput.value = '';
    }
  }

  loginTypeUsername.addEventListener('change', updateVisibility);
  loginTypeEmail.addEventListener('change', updateVisibility);

  // ---Password toggle---
  const toggle = document.getElementById('togglePassword');
  const eyeIcon = document.getElementById('eye-icon');
  const hiddenIcon = document.getElementById('hidden-icon');

  toggle?.addEventListener('change', () => {
    if (toggle.checked) {
      passwordInput.type = 'text';
      eyeIcon.style.display = 'none';
      hiddenIcon.style.display = 'block';
    } else {
      passwordInput.type = 'password';
      eyeIcon.style.display = 'block';
      hiddenIcon.style.display = 'none';
    }
  });

  // ---Clear errors on input---
  nicknameInput?.addEventListener('input', () => {
    nicknameError.textContent = '';
  });
  emailInput?.addEventListener('input', () => {
    emailError.textContent = '';
  });
  passwordInput?.addEventListener('input', () => {
    passwordError.textContent = '';
  });

  // ---Reset button---
  form.querySelector('button[type="reset"]').addEventListener('click', () => {
    nicknameError.textContent = '';
    emailError.textContent = '';
    passwordError.textContent = '';

    passwordInput.type = 'password';
    eyeIcon.style.display = 'block';
    hiddenIcon.style.display = 'none';
    if (toggle) toggle.checked = false;

    loginTypeUsername.checked = true;
    updateVisibility();
  });

  // ---Submit---
  form.addEventListener('submit', async (e) => {
    e.preventDefault();

    nicknameError.textContent = '';
    emailError.textContent = '';
    passwordError.textContent = '';

    const useEmail = loginTypeEmail.checked;
    const nickname = nicknameInput.value.trim();
    const email = emailInput.value.trim();
    const password = passwordInput.value.trim();

    // Client-side validation
    if (!password) {
      passwordError.textContent = 'Please enter your password.';
      passwordInput.focus();
      return;
    }
    if (useEmail && !email) {
      emailError.textContent = 'Please enter your email address.';
      emailInput.focus();
      return;
    }
    if (!useEmail && !nickname) {
      nicknameError.textContent = 'Please enter your nickname.';
      nicknameInput.focus();
      return;
    }

    const submitBtn = form.querySelector('button[type="submit"]');
    submitBtn.disabled = true;
    submitBtn.textContent = 'Signing in...';

    try {
      if (useEmail) {
        await api.post('/login/email', { email, password });
      } else {
        // Backend /login/username handler reads field "username"
        await api.post('/login/username', { username: nickname, password });
      }
      //once the cookies are set, navigate to homepage, over there the middlware will get the user
      navigate('/');
    } catch (err) {
      const msg = err?.message || String(err);
      const lower = msg.toLowerCase();

      if (lower.includes('email')) {
        emailError.textContent = escapeHTML(msg);
        emailInput.focus();
      } else if (
        lower.includes('nickname') ||
        lower.includes('nick') ||
        lower.includes('username')
      ) {
        nicknameError.textContent = escapeHTML(msg);
        nicknameInput.focus();
      } else {
        passwordError.textContent = escapeHTML(msg);
        passwordInput.focus();
      }
    } finally {
      submitBtn.disabled = false;
      submitBtn.textContent = 'Sign In';
    }
  });

  // Ensure initial visibility is correct
  updateVisibility();
}
