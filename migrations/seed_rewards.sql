-- Seed data untuk rewards
-- Money Reward 1-5 sesuai requirement

INSERT INTO `rewards` (`name`, `omset_target`, `reward_desc`, `duration`, `is_accumulative`, `status`) VALUES
('Money Reward 1', 5000000.00, 'Rp500.000 (tunai)', 30, 0, 'Active'),
('Money Reward 2', 15000000.00, 'Ponsel/ Gadget (Rp1,5–2 juta)', 30, 0, 'Active'),
('Money Reward 3', 30000000.00, 'Kulkas kecil / TV (Rp2,5–3 juta)', 60, 0, 'Active'),
('Money Reward 4', 75000000.00, 'Uang tunai 10.000.000', 90, 1, 'Active'),
('Money Reward 5', 200000000.00, 'Rp20.000.000 (tunai)', 150, 1, 'Active');

