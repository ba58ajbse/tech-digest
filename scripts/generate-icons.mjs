import { writeFileSync } from 'node:fs';
import { deflateSync } from 'node:zlib';

const signature = Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]);

function crc32(buf) {
  let crc = 0xffffffff;
  for (let i = 0; i < buf.length; i += 1) {
    crc ^= buf[i];
    for (let j = 0; j < 8; j += 1) {
      const mask = -(crc & 1);
      crc = (crc >>> 1) ^ (0xedb88320 & mask);
    }
  }
  return (crc ^ 0xffffffff) >>> 0;
}

function chunk(type, data) {
  const typeBuf = Buffer.from(type, 'ascii');
  const length = Buffer.alloc(4);
  length.writeUInt32BE(data.length, 0);
  const crcBuf = Buffer.alloc(4);
  const crc = crc32(Buffer.concat([typeBuf, data]));
  crcBuf.writeUInt32BE(crc, 0);
  return Buffer.concat([length, typeBuf, data, crcBuf]);
}

function toRgba(hex) {
  const value = hex.replace('#', '');
  const r = parseInt(value.slice(0, 2), 16);
  const g = parseInt(value.slice(2, 4), 16);
  const b = parseInt(value.slice(4, 6), 16);
  return [r, g, b, 255];
}

function makeIcon(size, path) {
  const background = toRgba('#1d1c1a');
  const accent = toRgba('#d66b2b');
  const accent2 = toRgba('#e3a64c');
  const rowSize = size * 4 + 1;
  const raw = Buffer.alloc(rowSize * size);

  for (let y = 0; y < size; y += 1) {
    const rowStart = y * rowSize;
    raw[rowStart] = 0; // no filter

    for (let x = 0; x < size; x += 1) {
      const offset = rowStart + 1 + x * 4;
      const diagonal = x + y;
      const useAccent = diagonal < size * 0.65;
      const useAccent2 = diagonal < size * 0.35;
      const color = useAccent2 ? accent2 : useAccent ? accent : background;
      raw[offset] = color[0];
      raw[offset + 1] = color[1];
      raw[offset + 2] = color[2];
      raw[offset + 3] = color[3];
    }
  }

  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(size, 0);
  ihdr.writeUInt32BE(size, 4);
  ihdr[8] = 8; // bit depth
  ihdr[9] = 6; // color type RGBA
  ihdr[10] = 0;
  ihdr[11] = 0;
  ihdr[12] = 0;

  const compressed = deflateSync(raw);
  const png = Buffer.concat([
    signature,
    chunk('IHDR', ihdr),
    chunk('IDAT', compressed),
    chunk('IEND', Buffer.alloc(0))
  ]);

  writeFileSync(path, png);
}

makeIcon(192, new URL('../public/icons/icon-192.png', import.meta.url));
makeIcon(512, new URL('../public/icons/icon-512.png', import.meta.url));
makeIcon(180, new URL('../public/icons/apple-touch-icon.png', import.meta.url));
