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
	
create or replace view nlink as
	select page_from p0, page_to p1 from link where page_from <> page_to
	union
	select page_to p0, page_from p1 from link where page_from <> page_to;