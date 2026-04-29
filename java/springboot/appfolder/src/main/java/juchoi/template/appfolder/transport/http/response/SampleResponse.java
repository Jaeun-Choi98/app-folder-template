package juchoi.template.appfolder.transport.http.response;

import lombok.RequiredArgsConstructor;

@RequiredArgsConstructor
public class SampleResponse {
    private final String result;

    public String getResult() { return result; }
}
