-- Create storage bucket for job artifacts (PDFs, spreadsheets, etc.)
INSERT INTO storage.buckets (id, name, public)
VALUES ('job-artifacts', 'job-artifacts', true)
ON CONFLICT (id) DO NOTHING;

-- Allow public reads (artifacts are accessible via URL)
CREATE POLICY "Public read access for job artifacts"
ON storage.objects FOR SELECT
USING (bucket_id = 'job-artifacts');

-- Allow service role to insert artifacts
CREATE POLICY "Service role can upload job artifacts"
ON storage.objects FOR INSERT
WITH CHECK (bucket_id = 'job-artifacts');

-- Allow service role to delete artifacts
CREATE POLICY "Service role can delete job artifacts"
ON storage.objects FOR DELETE
USING (bucket_id = 'job-artifacts');
