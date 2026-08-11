-- Developer public profile fields.
-- Social links and avatar are already available via the linked users row
-- (avatar_url, display_name). These columns power the public developer
-- profile page and the "meet the developer" section on product pages.

ALTER TABLE developers ADD COLUMN bio TEXT NOT NULL DEFAULT '';
ALTER TABLE developers ADD COLUMN location TEXT NOT NULL DEFAULT '';
ALTER TABLE developers ADD COLUMN website TEXT NOT NULL DEFAULT '';