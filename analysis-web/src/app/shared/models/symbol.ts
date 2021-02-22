import { Currency } from "./currency";

export class Symbol {
    id: number;
    currency: Currency;
    historyId: number;
    isin: string;
    identifier: string;
    name: string;
    minimumOrderQuantity: number;
    marketName: string;
    marketHoursGmt: string
}

export class Symbols {
    Symbols: Symbol[];
}