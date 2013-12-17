create table page (
	id serial not null, 
	url text not null, 
	title text not null,
	"text" text not null,
	class integer,
	human_class integer,
	constraint pk_page primary key (id),
	constraint uk_page_url unique (url),
	constraint chk_page_human_class check (human_class is null or human_class = class)
);

create table link (
	id serial not null, 
	href text not null, 
	page_from integer not null, 
	page_to integer, 
	"text" text not null,
	constraint pk_link primary key (id),
	constraint fk_link_page_from foreign key (page_from) 
		references page (id) on update cascade on delete cascade, 
	constraint fk_link_page_to foreign key (page_to) 
		references page (id) on update cascade on delete cascade
);

create index idx_link_href on link(href);
create index idx_link_page_from on link(page_from);
create index idx_link_page_to on link(page_to);
	
create table nlink (
	id serial not null,
	link_id int not null,
	page_0 int not null,
	page_1 int not null,
	constraint pk_nlink primary key (id),
	constraint fk_nlink_link foreign key (link_id) 
		references link (id) on update cascade on delete cascade, 
	constraint fk_nlink_page_0 foreign key (page_0) 
		references page (id) on update cascade on delete cascade, 
	constraint fk_nlink_page_1 foreign key (page_1) 
		references page (id) on update cascade on delete cascade
);

create index idx_nlink_link_id on nlink(link_id);
create index idx_nlink_page_0 on nlink(page_0, page_1);
create index idx_nlink_page_1 on nlink(page_1, page_0);
	
insert into nlink (link_id, page_0, page_1) select id, page_from, page_to from link union select id, page_to, page_from from link;
