import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/internal/Observable';
import { PagedRequest } from 'src/app/shared/models/request';
import { Symbols } from 'src/app/shared/models/symbol';
import { ConfigService } from 'src/app/shared/services/config.service';

@Injectable({
  providedIn: 'root'
})
export class SymbolsService {

  constructor(private http: HttpClient, private configService: ConfigService) { }

  getPaged(req: PagedRequest): Observable<Symbols> {
    return this.http.post<Symbols>(this.configService.apiBaseUrl + 'symbols/paged', req);
  }
}
