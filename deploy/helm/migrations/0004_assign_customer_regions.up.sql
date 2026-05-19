-- Assign realistic regions to the seeded fixture customers so the
-- region-scoped authz demo is meaningful. search_customer/lookup_customer
-- gate region='atlantis' rows behind the customers:atlantis:read permission
-- (see internal/store/queries.go). 0002 seeds 100 customers BEFORE the
-- region column exists (added in 0003 with DEFAULT 'unknown'), so the
-- fixtures all start 'unknown'; distribute them deterministically by id and
-- reserve a clear 'atlantis' cohort (ids 10,20,...,100) for the gated-access
-- demonstration. Deterministic so the demo/walkthrough is reproducible.
UPDATE customers
SET region = CASE
    WHEN id % 10 = 0 THEN 'atlantis'
    WHEN id % 3  = 0 THEN 'asia-pacific'
    WHEN id % 3  = 1 THEN 'north-america'
    ELSE 'europe'
END
WHERE region = 'unknown';
