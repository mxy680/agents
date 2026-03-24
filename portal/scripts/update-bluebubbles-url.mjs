import { createClient } from '@supabase/supabase-js';
import crypto from 'crypto';

const newUrl = process.argv[2];
if (!newUrl) {
  console.error('Usage: node update-bluebubbles-url.mjs <new-url>');
  process.exit(1);
}

function encrypt(plaintext) {
  const key = Buffer.from(process.env.ENCRYPTION_MASTER_KEY, 'hex');
  const nonce = crypto.randomBytes(12);
  const cipher = crypto.createCipheriv('aes-256-gcm', key, nonce);
  const encrypted = Buffer.concat([cipher.update(plaintext, 'utf8'), cipher.final()]);
  const tag = cipher.getAuthTag();
  return Buffer.concat([nonce, encrypted, tag]);
}

function decrypt(buf) {
  const key = Buffer.from(process.env.ENCRYPTION_MASTER_KEY, 'hex');
  const nonce = buf.subarray(0, 12);
  const tag = buf.subarray(buf.length - 16);
  const ciphertext = buf.subarray(12, buf.length - 16);
  const decipher = crypto.createDecipheriv('aes-256-gcm', key, nonce);
  decipher.setAuthTag(tag);
  return Buffer.concat([decipher.update(ciphertext), decipher.final()]).toString('utf8');
}

const s = createClient(process.env.NEXT_PUBLIC_SUPABASE_URL, process.env.SUPABASE_SERVICE_ROLE_KEY);

const { data: row } = await s.from('user_integrations').select('id, credentials').eq('provider', 'bluebubbles').eq('status', 'active').limit(1).single();
if (!row) { console.error('No bluebubbles integration found'); process.exit(1); }

const hex = row.credentials.startsWith('\\x') ? row.credentials.slice(2) : row.credentials;
const creds = JSON.parse(decrypt(Buffer.from(hex, 'hex')));
const oldUrl = creds.url;
creds.url = newUrl.replace(/\/+$/, '');

const encrypted = encrypt(JSON.stringify(creds));
const { error } = await s.from('user_integrations').update({ credentials: `\\x${encrypted.toString('hex')}`, updated_at: new Date().toISOString() }).eq('id', row.id);
if (error) { console.error('DB error:', error); process.exit(1); }
console.log(`Updated: ${oldUrl} → ${creds.url}`);
