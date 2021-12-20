PROTO_DIR=gcgrpc
PROTO_BASENAME=gcgrpc
PROTO_TARGET=$(PROTO_DIR)/$(PROTO_BASENAME).pb.go

all: $(PROTO_TARGET) 

$(PROTO_TARGET):
	protoc -I $(PROTO_DIR) --go_out=plugins=grpc:$(PROTO_DIR) $(PROTO_DIR)/$(PROTO_BASENAME).proto
	cp $(PROTO_DIR)/gc/$(PROTO_TARGET) $(PROTO_DIR)
	rm -rf $(PROTO_DIR)/gc

test:
	go test

clean:
	-rm -rf $(PROTO_TARGET)
