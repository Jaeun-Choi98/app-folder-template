package juchoi.template.appfolder.transport.http.response;

import lombok.RequiredArgsConstructor;

// Mirrors Go's response/base_response.go
@RequiredArgsConstructor
public class BaseResponse<T> {

    private final boolean success;
    private final String message;
    private final T data;

    public static <T> BaseResponse<T> ok(T data) {
        return new BaseResponse<>(true, "ok", data);
    }

    public static <T> BaseResponse<T> fail(String message) {
        return new BaseResponse<>(false, message, null);
    }

    public boolean isSuccess() { return success; }
    public String getMessage() { return message; }
    public T getData()         { return data; }
}
