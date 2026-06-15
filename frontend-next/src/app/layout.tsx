import type { Metadata } from 'next';
import { Rubik, Poppins } from 'next/font/google';
import '../styles/base.css';
import '../styles/layout.css';
import '../styles/components.css';
import '../styles/navbar.css';
import '../styles/activity.css';
import '../styles/category.css';
import '../styles/chat.css';
import '../styles/create-post.css';
import '../styles/filter-pagination.css';
import '../styles/signup-login.css';
import '../styles/spa.css';
import '../styles/topic.css';

// Configure Rubik font
const rubik = Rubik({
  subsets: ['latin'],
  weight: ['300', '400', '500', '600', '700', '800', '900'],
  style: ['normal', 'italic'],
  variable: '--font-rubik',
  display: 'swap',
});

// Configure Poppins font
const poppins = Poppins({
  subsets: ['latin'],
  weight: ['100', '200', '300', '400', '500', '600', '700', '800', '900'],
  style: ['normal', 'italic'],
  variable: '--font-poppins',
  display: 'swap',
});

export const metadata: Metadata = {
  title: 'Social Network',
  description: 'A real-time social networking platform',
  icons: {
    icon: '/images/icons/logo-icon.png',
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className={`${rubik.variable} ${poppins.variable}`}>
      <body>{children}</body>
    </html>
  );
}
