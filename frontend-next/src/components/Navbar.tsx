'use client';

/**
 * components/Navbar.tsx
 *
 * Mirrors renderNavbar() + buildNavbarHTML() from navbar.js.
 *
 * Omits notifications and chat for now — those will be added as separate
 * feature branches once the base migration is complete.
 *
 * Uses Next.js <Link> instead of <a data-link> — the router handles
 * client-side navigation automatically, no interceptor needed.
 */

import Link from 'next/link';
import Image from 'next/image';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect, useRef, useState } from 'react';
import { useAuth } from '@/context/AuthContext';
import { api } from '@/lib/api';

export default function Navbar() {
  const { user, clearUser } = useAuth();
  const pathname = usePathname();

  if (!user) return <GuestNav pathname={pathname} />;
  return <LoggedInNav user={user} pathname={pathname} clearUser={clearUser} />;
}

// ─── Guest navbar ─────────────────────────────────────────────────────────────

function GuestNav({ pathname }: { pathname: string }) {
  return (
    <header>
      <nav className="navbar">
        <div className="nav-container">
          <div className="logo">
            <Link className="logo-link" href="/">
              <div className="logo-icon">
                <Image src="/images/icons/logo-icon.png" alt="Logo Icon" width={32} height={32} />
              </div>
              <span className="logo-title">SocialNet</span>
            </Link>
          </div>
          <ul className="nav-links">
            <li className={`nav-link${pathname === '/login' ? ' active' : ''}`}>
              <Link href="/login">Login</Link>
            </li>
            <li className={`nav-link${pathname === '/register' ? ' active' : ''}`}>
              <Link href="/register">Register</Link>
            </li>
          </ul>
        </div>
      </nav>
    </header>
  );
}

// ─── Logged-in navbar ─────────────────────────────────────────────────────────

interface LoggedInNavProps {
  user: NonNullable<ReturnType<typeof useAuth>['user']>;
  pathname: string;
  clearUser: () => void;
}

function LoggedInNav({ user, pathname, clearUser }: LoggedInNavProps) {
  const router = useRouter();
  const [menuOpen, setMenuOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);

  const avatarSrc = user.avatar_url || '/images/user-avatar.png';

  const username = user.username || '';

  // Close dropdown on outside click — mirrors initUserMenu()
  useEffect(() => {
    function handleOutsideClick(e: MouseEvent) {
      if (
        menuRef.current &&
        triggerRef.current &&
        !menuRef.current.contains(e.target as Node) &&
        !triggerRef.current.contains(e.target as Node)
      ) {
        setMenuOpen(false);
      }
    }
    document.addEventListener('click', handleOutsideClick);
    return () => document.removeEventListener('click', handleOutsideClick);
  }, []);

  async function handleLogout() {
    try {
      await api.post('/logout');
    } catch (err) {
      console.warn('Backend logout failed, clearing client state anyway:', err);
    }
    clearUser();
    router.push('/login');
  }

  // const navLinks = [
  //   { href: '/topics', label: 'Topics' },
  //   { href: '/activity', label: 'Activity' },
  // ];

  return (
    <header>
      <nav className="navbar">
        <div className="nav-container">
          {/* Logo */}
          <div className="logo">
            <Link className="logo-link" href="/">
              <div className="logo-icon">
                <Image src="/images/icons/logo-icon.png" alt="Logo Icon" width={32} height={32} />
              </div>
              <span className="logo-title">SocialNet</span>
            </Link>
          </div>

          {/* Right side */}
          <div className="welcome-box">
            <div className="welcome-user-box">
              <div className="welcome-user-text">
                <span className="welcome-user">Welcome,</span>
                <span className="welcome-user-name">{username}</span>
              </div>

              {/* Avatar + dropdown */}
              <div className="user-menu-wrapper">
                <button
                  ref={triggerRef}
                  type="button"
                  className="user-avatar-link user-avatar-trigger"
                  aria-label="Open account menu"
                  aria-expanded={menuOpen}
                  onClick={(e) => {
                    e.stopPropagation();
                    setMenuOpen((prev) => !prev);
                  }}
                >
                  <span className="user-avatar">
                    <Image
                      src={avatarSrc}
                      alt="User Avatar"
                      width={36}
                      height={36}
                      onError={(e) => {
                        (e.target as HTMLImageElement).src = '/images/user-avatar.png';
                      }}
                    />
                  </span>
                </button>

                <div
                  ref={menuRef}
                  className="user-menu-dropdown"
                  style={{ display: menuOpen ? 'block' : 'none' }}
                >
                  <Link
                    href="/activity"
                    className="user-menu-item"
                    onClick={() => setMenuOpen(false)}
                  >
                    Profile
                  </Link>
                  <Link
                    href="/account/settings"
                    className="user-menu-item"
                    onClick={() => setMenuOpen(false)}
                  >
                    Account Settings
                  </Link>
                </div>
              </div>
            </div>

            {/* Nav links */}
            <ul className="nav-links">
              {/* {navLinks.map(({ href, label }) => (
                <li key={href} className={`nav-link${pathname === href ? ' active' : ''}`}>
                  <Link href={href}>{label}</Link>
                </li>
              ))} */}

              <li className="nav-link nav-link-create">
                <Link href="/topics/create">New Post</Link>
              </li>

              <li className="nav-link">
                {/* Logout is now a button to avoid a full page load */}
                <button
                  type="button"
                  onClick={handleLogout}
                  style={{
                    background: 'none',
                    border: 'none',
                    cursor: 'pointer',
                    font: 'inherit',
                    color: 'inherit',
                    padding: 0,
                  }}
                >
                  Logout
                </button>
              </li>
            </ul>
          </div>
        </div>
      </nav>
    </header>
  );
}
