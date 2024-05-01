# eMPI: Message Passing Interface

eMPI contains Go wrappers around the MPI message passing interface for distributed memory computation, in the `empi/mpi` package.  This has no other dependencies and uses code generation to provide support for all Go types.

You must set the `mpi` build tag to actually have it build using the mpi library -- the default is to build a dummy version that has 1 proc of rank 0 always, and nop versions of all the methods.

```bash
$ go build -tags mpi
```

The `empi/empi` package has methods to support use of MPI in emergent simulations:

* Gathering `table.Table` and `tensor.Tensor` data across processors.

* `AllocN` allocates n items to process across mpi processors.

## Development

After updating any of the template files, you need to update the generated go files like so:
```bash
cd mpi
go install github.com/apache/arrow/go/arrow/_tools/tmpl
make generate
```
