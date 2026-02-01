import type { Metadata } from 'next';
import './globals.css';
import PwaRegister from '@/components/pwa-register';

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
    <html lang="ja">
      <body>
        <PwaRegister />
        {children}
      </body>
    </html>
  );
}
