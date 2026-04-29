package juchoi.template.appfolder.service.api;

import juchoi.template.appfolder.db.handler.DbHandler;
import juchoi.template.appfolder.infra.cache.Cache;
import juchoi.template.appfolder.service.ApiService;
import juchoi.template.appfolder.transport.http.request.SampleRequest;
import juchoi.template.appfolder.transport.http.response.SampleResponse;
import lombok.RequiredArgsConstructor;

// ServiceConfig에서 @Bean으로 등록
@RequiredArgsConstructor
public class ApiServiceImpl implements ApiService {

    private final DbHandler dbHandler;
    private final Cache cache;

    @Override
    public String test() {
        return "hello world";
    }

    @Override
    public SampleResponse handleSample(SampleRequest request) {
        return new SampleResponse("processed: " + request.getName());
    }
}
