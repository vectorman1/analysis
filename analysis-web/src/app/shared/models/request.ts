export class PagedRequest {
    filter: Filter;
}

export class Filter {
    pageSize: number;
    pageNumber: number;
    order: string;
    ascending: boolean;
}