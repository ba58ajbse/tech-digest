import type { Metadata } from 'next';
import { Fraunces, Space_Grotesk } from 'next/font/google';
import './globals.css';

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
  description: 'Daily tech digest with Japanese summaries.'
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja" className={`${spaceGrotesk.variable} ${fraunces.variable}`}>
      <body>
        {children}
      </body>
    </html>
  );
}
