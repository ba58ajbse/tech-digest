import type { Metadata } from 'next';
import { Fraunces, Space_Grotesk } from 'next/font/google';
import './globals.css';
import PwaRegister from '@/components/pwa-register';

const spaceGrotesk = Space_Grotesk({
  subsets: ['latin'],
  variable: '--font-sans',
  display: 'swap'
});

const fraunces = Fraunces({
  subsets: ['latin'],
  variable: '--font-display',
  display: 'swap'
});

export const metadata: Metadata = {
  title: 'Tech Digest',
  description: 'Daily tech digest with Japanese summaries.',
  manifest: '/manifest.webmanifest',
  themeColor: '#1d1c1a',
  icons: {
    icon: ['/icons/icon-192.png', '/icons/icon-512.png'],
    apple: ['/icons/apple-touch-icon.png']
  }
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja" className={`${spaceGrotesk.variable} ${fraunces.variable}`}>
      <body>
        <PwaRegister />
        {children}
      </body>
    </html>
  );
}
