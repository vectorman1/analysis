import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import jwt_decode from 'jwt-decode';
import { catchError } from 'rxjs/operators';
import { ConfigService } from 'src/app/shared/services/config.service';
import { RegisterResponse } from '../models/register.response';
import { User } from '../models/user';

@Injectable({
  providedIn: 'root'
})
export class UserService {

  constructor(private http: HttpClient, private configService: ConfigService) { 

  }

  public getStoredUser(): User {
	let token = localStorage.getItem('id_token')

    if (!token) {
      return null;
	}
	
	let decodedToken = jwt_decode(token) as User;

	return decodedToken;
  }

  public logout() {
	  localStorage.removeItem('id_token')
  }

  public register(userRegistration: any): Observable<RegisterResponse> {
	  return this.http
		  .post<RegisterResponse>(this.configService.apiBaseUrl + 'user/register', userRegistration)
		  .pipe(
			  catchError()
		  )
  }
}
