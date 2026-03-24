import { createClient } from '@supabase/supabase-js';
import crypto from 'crypto';

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
const { data } = await s.from('user_integrations').select('credentials').eq('provider', 'streeteasy').eq('status', 'active').limit(1).single();
if (!data) { console.log('No streeteasy integration found'); process.exit(0); }
const hex = data.credentials.startsWith('\\x') ? data.credentials.slice(2) : data.credentials;
const creds = JSON.parse(decrypt(Buffer.from(hex, 'hex')));
const cookies = creds.all_cookies || '';
console.log('Cookie count:', cookies.split('; ').length);
console.log('Cookie names:', cookies.split('; ').map(c => c.split('=')[0]).join(', '));
console.log('\nHas _px3:', cookies.includes('_px3'));
console.log('Has _pxvid:', cookies.includes('_pxvid'));
console.log('Has pxcts:', cookies.includes('pxcts'));
