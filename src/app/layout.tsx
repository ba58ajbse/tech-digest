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

const themeScript = `(function(){try{var t=localStorage.getItem('theme');if(!t){t=matchMedia('(prefers-color-scheme:light)').matches?'light':'dark'}document.documentElement.setAttribute('data-theme',t);var c=t==='light'?'#f5f7fa':'#1d1c1a';var m=document.querySelector('meta[name="theme-color"]');if(m){m.setAttribute('content',c)}}catch(e){}})()`;

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja" data-theme="dark" suppressHydrationWarning>
      <head>
        <script dangerouslySetInnerHTML={{ __html: themeScript }} />
      </head>
      <body>
        <PwaRegister />
        {children}
      </body>
    </html>
  );
}
