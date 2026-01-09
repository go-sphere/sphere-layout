# typescript-api

### Adapter for PureAdmin

```typescript
import { Api, type GinxErrorResponse, HttpClient } from "@/api/swagger/Api";
import { http } from "@/utils/http";
import type { AxiosResponse } from "axios";
import type { PureHttpError } from "@/utils/http/types";

type UnwrapResponse<T> =
    T extends Promise<AxiosResponse<infer R>> ? Promise<R> : T;

type TransformFunction<T> = T extends (...args: infer Args) => infer Return
    ? (...args: Args) => UnwrapResponse<Return>
    : never;

type UnwrappedApiClient<T> = {
    [K in keyof Api<T>["api"]]: TransformFunction<Api<T>["api"][K]>;
};

type AdapterAPI<T> = UnwrappedApiClient<T>;

function createNewAPI(): AdapterAPI<unknown> {
    const client = new HttpClient<unknown>();
    client.instance = (http.constructor as any).axiosInstance;

    client.instance.interceptors.response.use(
        resp => resp,
        (err: PureHttpError) => {
            if (!err.isCancelRequest && err.response?.data) {
                const { code, message } = err.response.data as GinxErrorResponse;
                Object.assign(err, {
                    message: message || err.message,
                    errCode: code
                });
            }
            return Promise.reject(err);
        }
    );

    const api = new Api<unknown>(client);
    return api.api as AdapterAPI<unknown>;
}

export const API = createNewAPI();
```