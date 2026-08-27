-- 1. เพิ่มแบรนด์ระดับโลกเพิ่มเติม (Razer, Corsair, ASUS ROG, SteelSeries)
INSERT INTO brands (id, name, logo_url, description) VALUES
(5, 'Razer', 'https://images.unsplash.com/photo-1542751371-adc38448a05e?w=100', 'For Gamers. By Gamers.'),
(6, 'Corsair', 'https://images.unsplash.com/photo-1587202372775-e229f172b9d7?w=100', 'High-Performance PC Gaming Gear'),
(7, 'ASUS ROG', 'https://images.unsplash.com/photo-1593642632823-8f785ba67e45?w=100', 'Republic of Gamers'),
(8, 'SteelSeries', 'https://images.unsplash.com/photo-1598550476439-6847785fcea6?w=100', 'Winning is Everything')
ON CONFLICT (id) DO NOTHING;

SELECT setval('brands_id_seq', (SELECT MAX(id) FROM brands));

-- 2. ใส่สินค้าคุณภาพสูงให้ครบ 50 รายการ (IDs 5 - 50)
INSERT INTO products (id, name, description, price, stock, category_id, brand_id, image_url, sku, rating, reviews_count, specs) VALUES
-- Keyboards
(5, 'Keychron Q1 Pro Wireless Custom Keyboard', 'QMK/VIA Wireless Custom Mechanical Keyboard with Knob', 7290.00, 8, 1, 1, 'https://images.unsplash.com/photo-1595225476474-87563907a212?w=800&q=80', 'KB-KEY-005', 4.9, 88, '{"layout": "75%", "case": "Full CNC Aluminum", "connectivity": "Bluetooth 5.1 / Type-C", "hot_swappable": true}'),
(6, 'Keychron K3 Ultra-Slim Mechanical Keyboard', 'Ultra-slim wireless mechanical keyboard with low profile switches', 3490.00, 15, 2, 1, 'https://images.unsplash.com/photo-1618384887929-16ec33fab9ef?w=800&q=80', 'KB-KEY-006', 4.7, 120, '{"layout": "75%", "profile": "Low Profile Optical", "battery": "1550 mAh", "weight": "396g"}'),
(7, 'Razer BlackWidow V4 Pro', 'Mechanical Gaming Keyboard with Razer Chroma RGB and Command Dial', 8990.00, 5, 1, 5, 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&q=80', 'KB-RAZ-007', 4.8, 64, '{"switch": "Razer Green Clicky", "polling_rate": "8000Hz", "wrist_rest": "Plush Leatherette with Underglow"}'),
(8, 'Razer Huntsman Mini 60%', 'Optical Gaming Keyboard with Linear Optical Switches', 3990.00, 12, 1, 5, 'https://images.unsplash.com/photo-1595225476474-87563907a212?w=800&q=80', 'KB-RAZ-008', 4.6, 95, '{"layout": "60%", "switch": "Razer Linear Optical Gen-2", "keycaps": "Doubleshot PBT"}'),
(9, 'Corsair K70 RGB PRO', 'Mechanical Gaming Keyboard with AXON Hyper-Processing Technology', 5990.00, 10, 1, 6, 'https://images.unsplash.com/photo-1618384887929-16ec33fab9ef?w=800&q=80', 'KB-COR-009', 4.7, 52, '{"switch": "CHERRY MX Red", "frame": "Brushed Aluminum", "polling_rate": "8000Hz"}'),
(10, 'SteelSeries Apex Pro TKL', 'World Fastest Mechanical Keyboard with OmniPoint 2.0 Adjustable Switches', 8290.00, 6, 1, 8, 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&q=80', 'KB-STE-010', 5.0, 110, '{"switch": "OmniPoint 2.0 Magnetic", "actuation": "0.2mm - 3.8mm", "display": "OLED Smart Display"}'),

-- Mice & Trackpads
(11, 'Logitech MX Master 3S', 'Advanced Performance Wireless Ergonomic Mouse for Productivity', 3890.00, 25, 2, 2, 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=800&q=80', 'MS-LOGI-011', 4.9, 310, '{"sensor": "Darkfield 8000 DPI", "scroll": "MagSpeed Electromagnetic", "battery": "70 Days", "clicks": "Quiet Clicks"}'),
(12, 'Logitech G502 X PLUS', 'LIGHTSPEED Wireless RGB Gaming Mouse with LIGHTFORCE Hybrid Switches', 5290.00, 14, 1, 2, 'https://images.unsplash.com/photo-1527864550417-7fd91fc51a46?w=800&q=80', 'MS-LOGI-012', 4.8, 145, '{"sensor": "HERO 25K", "weight": "106g", "battery": "120 Hours", "buttons": "13 Programmable"}'),
(13, 'Razer DeathAdder V3 Pro', 'Ultra-lightweight Wireless Ergonomic Esports Gaming Mouse', 5490.00, 9, 1, 5, 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=800&q=80', 'MS-RAZ-013', 4.9, 175, '{"weight": "63g", "sensor": "Focus Pro 30K Optical", "switches": "Optical Gen-3", "battery": "90 Hours"}'),
(14, 'Razer Viper V2 Pro', 'Ultra-lightweight Symmetrical Wireless Gaming Mouse (58g)', 4990.00, 0, 1, 5, 'https://images.unsplash.com/photo-1527864550417-7fd91fc51a46?w=800&q=80', 'MS-RAZ-014', 4.7, 83, '{"weight": "58g", "connectivity": "HyperSpeed Wireless", "dpi": "30,000 DPI", "stock_status": "Out of Stock"}'),
(15, 'SteelSeries Aerox 3 Wireless', 'Ultra Lightweight 68g Water Resistant Gaming Mouse', 2990.00, 18, 1, 8, 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=800&q=80', 'MS-STE-015', 4.5, 62, '{"protection": "IP54 AquaBarrier", "battery": "200 Hours", "weight": "68g", "sensor": "TrueMove Air"}'),
(16, 'Corsair Dark Core RGB Pro', 'Wireless Gaming Mouse with Slipstream Technology', 3290.00, 11, 1, 6, 'https://images.unsplash.com/photo-1527864550417-7fd91fc51a46?w=800&q=80', 'MS-COR-016', 4.4, 48, '{"sensor": "Custom PixArt PAW3392 18K", "charging": "Qi Wireless Compatible", "polling": "2000Hz"}'),

-- Headsets & Audio
(17, 'HyperX Cloud Alpha Wireless', 'Gaming Headset with Massive 300 Hours Battery Life', 6990.00, 12, 1, 3, 'https://images.unsplash.com/photo-1546435770-a3e426bf472b?w=800&q=80', 'HS-HYP-017', 5.0, 190, '{"battery": "Up to 300 Hours", "driver": "Dual Chamber 50mm", "audio": "DTS Headphone:X Spatial"}'),
(18, 'HyperX Cloud III Wired', 'Comfortable Esports Gaming Headset with Angled 53mm Drivers', 3190.00, 20, 1, 3, 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=800&q=80', 'HS-HYP-018', 4.8, 115, '{"microphone": "Ultra-clear 10mm with Mesh Filter", "cushion": "HyperX Signature Memory Foam"}'),
(19, 'Razer BlackShark V2 Pro (2023)', 'Wireless Esports Headset with HyperClear Super Wideband Mic', 7590.00, 7, 1, 5, 'https://images.unsplash.com/photo-1546435770-a3e426bf472b?w=800&q=80', 'HS-RAZ-019', 4.9, 130, '{"mic": "Razer HyperClear Super Wideband", "drivers": "TriForce Titanium 50mm", "battery": "70 Hours"}'),
(20, 'SteelSeries Arctis Nova Pro Wireless', 'Almighty Audio System with Active Noise Cancellation & Dual Battery', 13900.00, 4, 1, 8, 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=800&q=80', 'HS-STE-020', 5.0, 95, '{"anc": "4-mic Hybrid Active Noise Cancellation", "system": "Infinity Power Dual-Battery Hot Swap"}'),
(21, 'Logitech G PRO X 2 LIGHTSPEED', 'Wireless Gaming Headset with Revolutionary Graphene Drivers', 8990.00, 10, 1, 2, 'https://images.unsplash.com/photo-1546435770-a3e426bf472b?w=800&q=80', 'HS-LOG-021', 4.8, 72, '{"driver": "50mm Graphene Audio Drivers", "wireless": "LIGHTSPEED + Bluetooth + 3.5mm", "battery": "50 Hours"}'),

-- Displays & Monitors
(22, 'Alienware 27 Gaming Monitor (AW2723DF)', '280Hz Fast IPS 1ms QHD Gaming Monitor', 23900.00, 5, 1, 4, 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=800&q=80', 'MN-ALI-022', 4.9, 44, '{"panel": "Fast IPS 27-inch", "resolution": "QHD 2560 x 1440", "refresh": "280Hz OC", "hdr": "VESA DisplayHDR 600"}'),
(23, 'ASUS ROG Swift OLED PG27AQDM', '27-inch 240Hz 0.03ms 1440p OLED Gaming Monitor', 37900.00, 3, 1, 7, 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=800&q=80', 'MN-ROG-023', 5.0, 68, '{"panel": "OLED 27-inch", "response": "0.03ms GtG", "gamut": "DCI-P3 99%", "cooling": "Custom Heatsink"}'),
(24, 'ASUS ROG Strix XG32UQ', '32-inch 4K 160Hz HDMI 2.1 Fast IPS Gaming Monitor', 32900.00, 4, 1, 7, 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=800&q=80', 'MN-ROG-024', 4.8, 35, '{"resolution": "4K UHD 3840 x 2160", "hdmi": "Dual HDMI 2.1 Native", "refresh": "160Hz Fast IPS"}'),
(25, 'Dell UltraSharp 27 4K USB-C Hub Monitor (U2723QE)', 'Professional Creator Display with IPS Black Technology', 22500.00, 8, 2, 4, 'https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?w=800&q=80', 'MN-DEL-025', 4.9, 140, '{"panel": "IPS Black 2000:1 Contrast", "resolution": "4K UHD 3840 x 2160", "power_delivery": "90W USB-C PD Hub"}'),

-- Microphones & Streaming
(26, 'HyperX QuadCast S', 'USB Condenser Microphone with Dynamic RGB Lighting & Shock Mount', 5390.00, 16, 1, 3, 'https://images.unsplash.com/photo-1590658268037-6bf12165a8df?w=800&q=80', 'MC-HYP-026', 4.9, 210, '{"patterns": "Stereo, Omnidirectional, Cardioid, Bidirectional", "mount": "Built-in Anti-Vibration Shock Mount"}'),
(27, 'Razer Seiren V2 Pro', 'Professional Dynamic USB Microphone for Streamers', 4990.00, 12, 1, 5, 'https://images.unsplash.com/photo-1590658268037-6bf12165a8df?w=800&q=80', 'MC-RAZ-027', 4.7, 58, '{"capsule": "30mm Dynamic Capsule", "filter": "High Pass Filter & Analog Gain Limiter"}'),
(28, 'Logitech Blue Yeti X', 'Professional Multi-Pattern USB Mic with High-Res LED Metering', 5990.00, 9, 2, 2, 'https://images.unsplash.com/photo-1590658268037-6bf12165a8df?w=800&q=80', 'MC-LOG-028', 4.8, 180, '{"array": "Custom 4-Capsule Condenser Array", "meter": "11-Segment LED Metering", "effects": "Blue VO!CE Broadcast Filters"}'),
(29, 'Elgato Stream Deck MK.2', '15 Customizable LCD Keys for Studio Control', 5490.00, 14, 2, 6, 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&q=80', 'SD-ELG-029', 4.9, 160, '{"keys": "15 Tactile LCD Keys", "stand": "45-Degree Desktop Stand", "faceplate": "Detachable"}'),

-- Webcams
(30, 'Logitech Brio 4K Ultra HD Webcam', 'Ultra 4K HD Video Collaboration & Streaming Webcam with HDR', 6990.00, 10, 2, 2, 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&q=80', 'WC-LOG-030', 4.8, 220, '{"resolution": "4K Ultra HD @ 30fps / 1080p @ 60fps", "hdr": "RightLight 3 with HDR", "fov": "90° / 78° / 65°"}'),
(31, 'Razer Kiyo Pro Ultra', 'Largest Sensor in a Webcam for 4K 30FPS DSLR-like Quality', 11900.00, 4, 1, 5, 'https://images.unsplash.com/photo-1587829741301-dc798b83add3?w=800&q=80', 'WC-RAZ-031', 4.9, 42, '{"sensor": "Sony 1/1.2 STARVIS 2", "aperture": "F/1.7", "hdr": "True 4K 30FPS HDR"}'),

-- Ergonomic Office Gear
(32, 'Logitech Wave Keys Wireless Ergonomic Keyboard', 'Cushioned Palm Rest Ergonomic Keyboard for Day-Long Comfort', 2690.00, 22, 2, 2, 'https://images.unsplash.com/photo-1618384887929-16ec33fab9ef?w=800&q=80', 'KB-LOG-032', 4.7, 90, '{"design": "Wave Shape Compact Design", "palm_rest": "3-Layer Deep Memory Foam", "battery": "Up to 3 Years"}'),
(33, 'Logitech Lift Vertical Ergonomic Mouse', 'Ergonomic Vertical Mouse Designed for Small to Medium Hands', 2490.00, 20, 2, 2, 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=800&q=80', 'MS-LOG-033', 4.8, 140, '{"angle": "57-Degree Optimum Posture Angle", "battery": "Up to 2 Years", "clicks": "Whisper Quiet Clicks"}'),
(34, 'Herman Miller x Logitech G Embody Gaming Chair', 'Ergonomically Enhanced for Gaming & Supreme Spinal Support', 45900.00, 2, 2, 2, 'https://images.unsplash.com/photo-1580481077195-c3f9a941e7fa?w=800&q=80', 'CH-HER-034', 5.0, 85, '{"technology": "BackFit Adjustment with Pixelated Support", "foam": "Copper-fused Cooling Foam", "warranty": "12 Years"}'),

-- Gaming Chairs & Desks
(35, 'Razer Iskur V2 Ergonomic Gaming Chair', 'Adaptive Lumbar Support Gaming Chair with EPU Grade Leather', 21900.00, 5, 1, 5, 'https://images.unsplash.com/photo-1580481077195-c3f9a941e7fa?w=800&q=80', 'CH-RAZ-035', 4.8, 39, '{"lumbar": "Adaptive 6D Swivel Lumbar Support", "upholstery": "EPU Grade Synthetic Leather", "armrests": "4D Armrests"}'),
(36, 'Secretlab TITAN Evo Gaming Desk Mat (Stealth)', 'Magnetic Leatherette Desk Mat with Cable Management Anchors', 2490.00, 18, 1, 6, 'https://images.unsplash.com/photo-1527864550417-7fd91fc51a46?w=800&q=80', 'MP-SEC-036', 4.8, 64, '{"material": "Secretlab NEOTM Hybrid Leatherette", "bottom": "Non-Slip Anti-Fray Magnetic Base"}'),

-- Laptops & PCs
(37, 'ASUS ROG Zephyrus G14 (2024)', 'OLED Ultra-Slim Gaming Laptop Ryzen 9 RTX 4070 32GB 1TB', 69900.00, 3, 1, 7, 'https://images.unsplash.com/photo-1603302576837-37561b2e2302?w=800&q=80', 'LP-ROG-037', 5.0, 32, '{"display": "14-inch 3K 120Hz OLED 0.2ms", "gpu": "NVIDIA GeForce RTX 4070 8GB", "weight": "1.50 kg"}'),
(38, 'Alienware m16 R2 Gaming Laptop', 'Core Ultra 7 RTX 4070 240Hz QHD+ Stealth Mode Gaming Laptop', 65900.00, 2, 1, 4, 'https://images.unsplash.com/photo-1603302576837-37561b2e2302?w=800&q=80', 'LP-ALI-038', 4.9, 21, '{"cpu": "Intel Core Ultra 7 155H", "display": "16-inch QHD+ 240Hz 3ms", "cooling": "Alienware Cryo-tech"}'),

-- Mousepads & Accessories
(39, 'SteelSeries QcK Heavy XXL Mousepad', 'Extra Thick Gaming Mousepad with Micro-Woven Cloth (900x400x4mm)', 1190.00, 30, 1, 8, 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=800&q=80', 'MP-STE-039', 4.9, 410, '{"size": "900 x 400 x 4 mm", "surface": "Exclusive QcK Micro-Woven Cloth", "base": "Extra Thick Non-Slip Rubber"}'),
(40, 'Razer Strider Chroma Extended Hybrid Mat', 'Hybrid Soft/Hard Gaming Mouse Mat with 19 Chroma RGB Zones', 4290.00, 8, 1, 5, 'https://images.unsplash.com/photo-1527864550417-7fd91fc51a46?w=800&q=80', 'MP-RAZ-040', 4.7, 51, '{"lighting": "19 Multi-Zone Razer Chroma RGB", "surface": "Polyester Hybrid Gliding Surface"}'),
(41, 'Corsair MM700 RGB Extended Cloth Mouse Pad', 'Spacious 930x400mm Surface with 360-Degree Dynamic 3-Zone RGB', 2190.00, 15, 1, 6, 'https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?w=800&q=80', 'MP-COR-041', 4.8, 85, '{"hub": "Built-in 2-Port USB Passthrough Hub", "dimensions": "930 x 400 x 4 mm"}'),
(42, 'Keychron Coiled Aviator Cable (Carbon Black)', 'Custom Coiled Keyboard Cable with Detachable 5-Pin Aviator Port', 890.00, 40, 1, 1, 'https://images.unsplash.com/photo-1595225476474-87563907a212?w=800&q=80', 'CB-KEY-042', 4.9, 132, '{"connector": "Type-C to Type-A", "length": "140 cm", "aviator": "Detachable 5-pin GX16"}'),
(43, 'Keychron Wooden Palm Rest for K3 / K7', 'Solid American Walnut Wood Ergonomic Palm Rest', 790.00, 25, 2, 1, 'https://images.unsplash.com/photo-1618384887929-16ec33fab9ef?w=800&q=80', 'PR-KEY-043', 4.8, 97, '{"wood": "Solid American Walnut Wood", "finish": "Smooth Satin Oil Finish"}'),

-- Controllers & Gaming Gear
(44, 'Xbox Wireless Controller (Carbon Black)', 'Precision Wireless Controller with Textured Grip & Hybrid D-Pad', 2290.00, 18, 1, 2, 'https://images.unsplash.com/photo-1600080972464-8e5f35f63d08?w=800&q=80', 'CT-MSF-044', 4.9, 280, '{"connectivity": "Xbox Wireless / Bluetooth / Type-C", "battery": "Up to 40 Hours"}'),
(45, 'Sony PlayStation DualSense Wireless Controller', 'Immersive Haptic Feedback & Dynamic Adaptive Triggers', 2390.00, 16, 1, 2, 'https://images.unsplash.com/photo-1600080972464-8e5f35f63d08?w=800&q=80', 'CT-SON-045', 4.9, 320, '{"feedback": "Dual Actuators Haptic Feedback", "triggers": "Dynamic Adaptive Triggers with Variable Resistance"}'),
(46, 'Razer Wolverine V2 Pro Wireless Pro Controller', 'HyperSpeed Wireless Esports Controller for PS5 and PC', 9990.00, 4, 1, 5, 'https://images.unsplash.com/photo-1600080972464-8e5f35f63d08?w=800&q=80', 'CT-RAZ-046', 4.7, 44, '{"switches": "Razer Mecha-Tactile Action Buttons", "dpad": "8-Way Microswitch D-Pad", "wireless": "HyperSpeed Wireless 2.4GHz"}'),

-- Audio & Earbuds
(47, 'Sony WH-1000XM5 Wireless Noise Canceling Headphones', 'Industry Leading Noise Canceling with Auto NC Optimizer', 13990.00, 8, 2, 2, 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=800&q=80', 'HD-SON-047', 4.9, 450, '{"processors": "Integrated Processor V1 & HD Noise Canceling Processor QN1", "battery": "30 Hours with Quick Charge"}'),
(48, 'Razer Hammerhead Pro HyperSpeed Earbuds', 'Cross-Platform True Wireless Earbuds with Low-Latency Dongle', 7290.00, 10, 1, 5, 'https://images.unsplash.com/photo-1590658268037-6bf12165a8df?w=800&q=80', 'EB-RAZ-048', 4.6, 68, '{"latency": "Sub-20ms 2.4GHz Wireless via Type-C Dongle", "anc": "Adjustable Hybrid Active Noise Cancellation"}'),
(49, 'Logitech Zone Wireless 2 Headset', 'AI-Powered Noise-Canceling Bluetooth Headset for Business', 9990.00, 6, 2, 2, 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=800&q=80', 'HS-LOG-049', 4.8, 35, '{"ai": "AI-Powered 2-Way Noise-Free Calling", "range": "Up to 50m Wireless Freedom", "mic": "Flip-to-Mute Mic"}'),
(50, 'SteelSeries Tusq In-Ear Mobile Gaming Headset', 'Dynamic Sound Drivers & Dual Microphone System for Gaming', 1490.00, 24, 1, 8, 'https://images.unsplash.com/photo-1590658268037-6bf12165a8df?w=800&q=80', 'EB-STE-050', 4.5, 82, '{"drivers": "Dynamic Composite Sound Drivers", "microphone": "Detachable Boom Mic + Built-in Mic"}')
ON CONFLICT (id) DO NOTHING;

-- 3. อัปเดต Sequence ID สูงสุด
SELECT setval('products_id_seq', (SELECT MAX(id) FROM products));
