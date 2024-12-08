SQL Script--
User Service--------
create table
  public.users (
    id serial not null,
    email character varying(255) not null,
    phone_number character varying(20) null,
    password_hash character varying(255) not null,
    membership_tier character varying(20) null default 'Basic'::character varying,
    created_at timestamp without time zone null default current_timestamp,
    constraint users_pkey primary key (id),
    constraint users_email_key unique (email)
  ) tablespace pg_default;


Vehicle Service--------
create table
  public.vehicles (
    id serial not null,
    model character varying(100) not null,
    type character varying(50) not null,
    license_plate character varying(20) not null,
    status character varying(20) not null default 'available'::character varying,
    location character varying(255) null,
    battery_level integer null,
    cleanliness_status character varying(20) null default 'clean'::character varying,
    created_at timestamp without time zone null default current_timestamp,
    last_status_update timestamp without time zone null default current_timestamp,
    hourly_rate numeric(10, 2) not null default 9.00,
    constraint vehicles_pkey primary key (id),
    constraint vehicles_license_plate_key unique (license_plate),
    constraint vehicles_battery_level_check check (
      (
        (battery_level >= 0)
        and (battery_level <= 100)
      )
    )
  ) tablespace pg_default;

create index if not exists idx_vehicles_status on public.vehicles using btree (status) tablespace pg_default;

(Insert Vehicle Data)
WITH inserted_vehicles AS (
    INSERT INTO vehicles (model, type, license_plate, status, location, battery_level, cleanliness_status) 
    VALUES
        ('Tesla Model 3', 'Electric Sedan', 'SGP1234A', 'available', 'Marina Bay Sands', 90, 'clean'),
        ('Tesla Model Y', 'Electric SUV', 'SGP5678B', 'available', 'East Coast Park', 85, 'clean'),
        ('Nissan Leaf', 'Electric Hatchback', 'SGP9012C', 'available', 'ION Orchard', 75, 'clean'),
        ('BYD Atto 3', 'Electric SUV', 'SGP3456D', 'maintenance', 'Bugis Junction', 30, 'needs_cleaning'),
        ('Tesla Model Y', 'Electric SUV', 'SGP7890E', 'available', 'JEM Jurong East', 95, 'clean'),
        ('Hyundai Kona Electric', 'Electric SUV', 'SGP2345F', 'available', 'Kallang Wave Mall', 88, 'clean'),
        ('Kia EV6', 'Electric Crossover', 'SGP6789G', 'available', 'Somerset 313', 92, 'clean'),
        ('BYD Seal', 'Electric Sedan', 'SGP0123H', 'charging', 'Tampines Mall', 15, 'clean'),
        ('MG4', 'Electric Hatchback', 'SGP4567J', 'available', 'AMK Hub', 87, 'clean'),
        ('Tesla Model 3', 'Electric Sedan', 'SGP8901K', 'available', 'Clementi Mall', 83, 'needs_cleaning')
    RETURNING id, model
)

create table
  public.bookings (
    id serial not null,
    user_id integer not null,
    vehicle_id integer not null,
    start_time timestamp without time zone not null,
    end_time timestamp without time zone not null,
    status character varying(20) not null default 'pending'::character varying,
    created_at timestamp without time zone null default current_timestamp,
    updated_at timestamp without time zone null default current_timestamp,
    total_cost numeric(10, 2) null,
    constraint bookings_pkey primary key (id),
    constraint bookings_vehicle_id_fkey foreign key (vehicle_id) references vehicles (id),
    constraint valid_time_range check ((end_time > start_time))
  ) tablespace pg_default;

create index if not exists idx_bookings_user_id on public.bookings using btree (user_id) tablespace pg_default;

create index if not exists idx_bookings_vehicle_id on public.bookings using btree (vehicle_id) tablespace pg_default;

create index if not exists idx_bookings_status on public.bookings using btree (status) tablespace pg_default;

create trigger create_invoice_after_booking
after insert on bookings for each row
execute function create_invoice_for_booking ();

create table
  public.pricing_tiers (
    id serial not null,
    name character varying(50) not null,
    hourly_rate numeric(10, 2) not null,
    discount numeric(5, 2) not null,
    created_at timestamp without time zone null default current_timestamp,
    constraint pricing_tiers_pkey primary key (id),
    constraint pricing_tiers_name_key unique (name),
    constraint valid_discount check (
      (
        (discount >= (0)::numeric)
        and (discount <= (100)::numeric)
      )
    )
  ) tablespace pg_default;

INSERT INTO public.pricing_tiers (name, hourly_rate, discount) VALUES
    ('Basic', 9.00, 0.00),    -- 0% discount
    ('Premium', 9.00, 11.11), -- $1 off = ~11.11% discount on $9
    ('VIP', 9.00, 22.22);     -- $2 off = ~22.22% discount on $9

Billing Service ----------
create table
  public.invoices (
    id serial not null,
    booking_id integer not null,
    user_id integer not null,
    amount numeric(10, 2) not null,
    discount_amount numeric(10, 2) not null default 0.00,
    final_amount numeric(10, 2) not null,
    status character varying(20) not null default 'pending'::character varying,
    created_at timestamp without time zone null default current_timestamp,
    updated_at timestamp without time zone null default current_timestamp,
    constraint invoices_pkey primary key (id),
    constraint invoices_booking_id_fkey foreign key (booking_id) references bookings (id),
    constraint invoices_user_id_fkey foreign key (user_id) references users (id)
  ) tablespace pg_default;

create index if not exists idx_invoices_user_id on public.invoices using btree (user_id) tablespace pg_default;

create index if not exists idx_invoices_status on public.invoices using btree (status) tablespace pg_default;
