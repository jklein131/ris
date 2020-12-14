create table "order"
(
	id serial not null
		constraint order_pk
			primary key,
	order_number integer not null,
	order_date timestamp not null,
	cancelled boolean default false not null
);

create table line_item
(
	id serial not null
		constraint line_item_pk
			primary key,
	sku varchar not null,
	rush boolean default false,
	order_id integer
		constraint line_item_order_id_fk
			references "order"
);

create table component
(
	id serial not null
		constraint component_pk
			primary key,
	line_item_id integer not null
		constraint component_line_item_id_fk
			references line_item,
	status varchar default 'Pending'::character varying not null,
	size varchar not null
);


select * from component, line_item, "order"
where component.line_item_id = line_item.id
 and line_item.order_id = "order".id
 and component.status = 'Pending'
 and  "order".cancelled = 'false'
ORDER BY line_item.rush DESC, "order".order_date ASC

insert into "order" (id, order_number, order_date, cancelled) values (1, 2562, '2020-10-13 04:27:30.000000', 'false');
insert into "order" (id, order_number, order_date, cancelled) values (2, 2563, '2020-10-14 09:14:30.000000', 'false');
insert into "order" (id, order_number, order_date, cancelled) values (3, 2564, '2020-10-14 10:10:10.000000', 'false');
insert into "order" (id, order_number, order_date, cancelled) values (4, 2565, '2020-10-15 10:10:32.100000', 'true');
insert into "order" (id, order_number, order_date, cancelled) values (5, 2566, '2020-10-16 10:27:30.000000', 'false');
insert into "order" (id, order_number, order_date, cancelled) values (6, 2567, '2020-10-17 19:25:00.000000', 'false');

insert into line_item (id, sku, rush, order_id) values (1, 'RC-0027-27', 'false', 1);
insert into line_item (id, sku, rush, order_id) values (2, 'RC-0030-57', 'false', 1);
insert into line_item (id, sku, rush, order_id) values (3, 'RC-DY25-27', 'false', 2);
insert into line_item (id, sku, rush, order_id) values (4, 'RS-MY18-35', 'true', 3);
insert into line_item (id, sku, rush, order_id) values (5, 'RS-BS73-57', 'false', 4);
insert into line_item (id, sku, rush, order_id) values (6, 'RS-DS55-27', 'true', 5);
insert into line_item (id, sku, rush, order_id) values (7, 'RC-IH17-35', 'false', 5);
insert into line_item (id, sku, rush, order_id) values (8, 'RS-AH27-27', 'false', 6);

insert into component (id, line_item_id, status, size) values (1, 1, 'Pending', '2.5x7');
insert into component (id, line_item_id, status, size) values (2, 2, 'Pending', '5x7');
insert into component (id, line_item_id, status, size) values (3, 3, 'Pending', '2.5x7');
insert into component (id, line_item_id, status, size) values (4, 4, 'Pending', '3x5');
insert into component (id, line_item_id, status, size) values (5, 5, 'Cancelled', '5x7');
insert into component (id, line_item_id, status, size) values (6, 6, 'Pending', '2.5x7');
insert into component (id, line_item_id, status, size) values (7, 7, 'Pending', '3x5');
insert into component (id, line_item_id, status, size) values (8, 8, 'Pending', '2.5x7');
