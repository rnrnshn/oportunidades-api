-- Store country names in Portuguese, matching the admin country dropdown and the public site.
UPDATE "opportunities" SET "country" = 'Moçambique' WHERE "country" IN ('Mozambique', 'Mocambique');
UPDATE "universities" SET "country" = 'Moçambique' WHERE "country" IN ('Mozambique', 'Mocambique');
