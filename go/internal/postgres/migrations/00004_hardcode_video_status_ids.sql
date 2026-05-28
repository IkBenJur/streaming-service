-- +goose Up

ALTER TABLE videos DROP CONSTRAINT videos_status_fkey;

UPDATE videos SET status = '00000000-0000-0000-0000-000000000001'
WHERE status = (SELECT id FROM video_statuses WHERE status = 'pending');

UPDATE videos SET status = '00000000-0000-0000-0000-000000000002'
WHERE status = (SELECT id FROM video_statuses WHERE status = 'processing');

UPDATE videos SET status = '00000000-0000-0000-0000-000000000003'
WHERE status = (SELECT id FROM video_statuses WHERE status = 'finished');

UPDATE videos SET status = '00000000-0000-0000-0000-000000000004'
WHERE status = (SELECT id FROM video_statuses WHERE status = 'failed');

UPDATE video_statuses SET id = '00000000-0000-0000-0000-000000000001' WHERE status = 'pending';
UPDATE video_statuses SET id = '00000000-0000-0000-0000-000000000002' WHERE status = 'processing';
UPDATE video_statuses SET id = '00000000-0000-0000-0000-000000000003' WHERE status = 'finished';
UPDATE video_statuses SET id = '00000000-0000-0000-0000-000000000004' WHERE status = 'failed';

ALTER TABLE videos ADD CONSTRAINT videos_status_fkey FOREIGN KEY (status) REFERENCES video_statuses(id);

-- +goose Down

ALTER TABLE videos DROP CONSTRAINT videos_status_fkey;

UPDATE video_statuses SET id = gen_random_uuid() WHERE status = 'pending';
UPDATE video_statuses SET id = gen_random_uuid() WHERE status = 'processing';
UPDATE video_statuses SET id = gen_random_uuid() WHERE status = 'finished';
UPDATE video_statuses SET id = gen_random_uuid() WHERE status = 'failed';

UPDATE videos v SET status = vs.id
FROM video_statuses vs
WHERE vs.status = 'pending' AND v.status = '00000000-0000-0000-0000-000000000001';

UPDATE videos v SET status = vs.id
FROM video_statuses vs
WHERE vs.status = 'processing' AND v.status = '00000000-0000-0000-0000-000000000002';

UPDATE videos v SET status = vs.id
FROM video_statuses vs
WHERE vs.status = 'finished' AND v.status = '00000000-0000-0000-0000-000000000003';

UPDATE videos v SET status = vs.id
FROM video_statuses vs
WHERE vs.status = 'failed' AND v.status = '00000000-0000-0000-0000-000000000004';

ALTER TABLE videos ADD CONSTRAINT videos_status_fkey FOREIGN KEY (status) REFERENCES video_statuses(id);
