import { HttpClient } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { PagedRequest } from "../models/request";
import { Symbols } from "../models/symbol";
import { ConfigService } from "./config.service";

@Injectable({
    providedIn: 'root'
})
export class SymbolService {
    constructor(private http: HttpClient, private configService: ConfigService) {}

    getPaged(req: PagedRequest): Observable<Symbols> {
        return this.http.post<Symbols>(this.configService.apiBaseUrl + 'symbols/paged', req);
    }
}