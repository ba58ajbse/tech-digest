import type { Metadata, Viewport } from 'next';
import './globals.css';
import PwaRegister from '@/components/pwa-register';

export const metadata: Metadata = {
  title: 'Tech Digest',
  description: 'Daily tech digest with Japanese summaries.',
  manifest: '/manifest.webmanifest',
  icons: {
    icon: ['/icons/icon-192.png', '/icons/icon-512.png'],
    apple: ['/icons/apple-touch-icon.png']
  }
};

export const viewport: Viewport = {
  themeColor: '#1d1c1a'
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <body>
        <PwaRegister />
        {children}
      </body>
    </html>
  );
}
