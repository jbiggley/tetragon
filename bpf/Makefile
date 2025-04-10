.PHONY: all clean
.SUFFIXES:

SHELL=/bin/bash # needed for the *.{o,ll,i,s} pattern in the clean target

CLANG  ?= clang
LLC    ?= llc

# Build the BPF programs for the detected architecture, default to x86, and
# allow easy overriding by using ?= for cross-compilation
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),x86_64)
	BPF_TARGET_ARCH ?= x86
endif
ifeq ($(UNAME_M),aarch64)
	BPF_TARGET_ARCH ?= arm64
endif
BPF_TARGET_ARCH ?= x86

ALIGNCHECKERDIR = alignchecker/
PROCESSDIR := process/
CGROUPDIR := cgroup/
BPFTESTDIR := test/

ALIGNCHECKER = bpf_alignchecker.o
PROCESS = bpf_execve_event.o bpf_execve_event_v53.o bpf_fork.o bpf_exit.o bpf_generic_kprobe.o \
	  bpf_generic_kprobe_v53.o bpf_generic_retkprobe.o bpf_generic_retkprobe_v53.o \
	  bpf_multi_kprobe_v53.o bpf_multi_retkprobe_v53.o \
	  bpf_generic_tracepoint.o bpf_generic_tracepoint_v53.o \
	  bpf_generic_uprobe.o bpf_generic_uprobe_v53.o \
	  bpf_execve_event_v61.o \
	  bpf_generic_kprobe_v61.o bpf_generic_retkprobe_v61.o \
	  bpf_generic_tracepoint_v61.o \
	  bpf_multi_kprobe_v61.o bpf_multi_retkprobe_v61.o \
	  bpf_generic_uprobe_v61.o \
	  bpf_loader.o

CGROUP = bpf_cgroup_mkdir.o bpf_cgroup_rmdir.o bpf_cgroup_release.o
BPFTEST = bpf_lseek.o bpf_globals.o

IDIR = ./include/
LIBBPF = ./libbpf/
LDIR = ./lib
DEPS = $(patsubst %,$(IDIR)/%,$(_DEPS))

FLAGS := -I. \
	 -Wall -Werror \
	 -Wno-address-of-packed-member -Wno-compare-distinct-pointer-types -Wno-unknown-warning-option \
	 -O2

CLANG_FLAGS += $(FLAGS) -I $(LIBBPF) -I $(IDIR) -I $(LDIR) -target bpf -emit-llvm -g -D__TARGET_ARCH_$(BPF_TARGET_ARCH) -fdebug-default-version=4
LLC_FLAGS   := -march=bpf -mcpu=v2 -mattr=dwarfris
LLC_FLAGS_ALU32 := -march=bpf -mcpu=v3 -mattr=dwarfris

OBJSDIR       := objs/
DEPSDIR       := deps/

TLSOBJ        := $(addprefix $(OBJSDIR),$(TLS))
NOPOBJ        := $(addprefix $(OBJSDIR),$(NOP))
PROCESSOBJ    := $(addprefix $(OBJSDIR),$(PROCESS))
CGROUPOBJ     := $(addprefix $(OBJSDIR),$(CGROUP))
TESTOBJ       := $(addprefix $(OBJSDIR),$(BPFTEST))
ALIGNCHECKEROBJ := $(addprefix $(OBJSDIR),$(ALIGNCHECKER))
OBJS          := $(PROCESSOBJ) $(CGROUPOBJ) $(TESTOBJ) $(NOPOBJ) $(ALIGNCHECKEROBJ)
LLOBJS        := $(patsubst $(OBJSDIR)%.o,$(OBJSDIR)%.ll,$(OBJS))
DEPS          := $(patsubst $(OBJSDIR)%.ll,$(DEPSDIR)%.d,$(LLOBJS))

all: $(OBJS) $(DEPS)

# NB: https://www.gnu.org/software/make/manual/html_node/Prerequisite-Types.html
$(OBJS): | $(OBJSDIR)
$(DEPS): | $(DEPSDIR)
$(LLOBJS): | $(OBJSDIR)

$(OBJSDIR):
	mkdir $(OBJSDIR)

$(DEPSDIR):
	mkdir $(DEPSDIR)

define DEFINE_VARIANT
VAR := $1
deps/bpf_execve_event_$$(VAR).d: process/bpf_execve_event.c
deps/bpf_generic_kprobe_$$(VAR).d: process/bpf_generic_kprobe.c
deps/bpf_generic_retkprobe_$$(VAR).d: process/bpf_generic_retkprobe.c
deps/bpf_multi_kprobe_$$(VAR).d: process/bpf_generic_kprobe.c
deps/bpf_multi_retkprobe_$$(VAR).d: process/bpf_generic_retkprobe.c
deps/bpf_generic_tracepoint_$$(VAR).d: process/bpf_generic_tracepoint.c
deps/bpf_generic_uprobe_$$(VAR).d: process/bpf_generic_uprobe.c
endef

# Generic build targets for each sub-dir

$(eval $(call DEFINE_VARIANT,v53))
$(eval $(call DEFINE_VARIANT,v61))

# ALIGNCHECKER
objs/%.ll: $(ALIGNCHECKERDIR)%.c
	$(CLANG) $(CLANG_FLAGS) -c $< -o $@

$(DEPSDIR)%.d: $(ALIGNCHECKERDIR)%.c
	$(CLANG) $(CLANG_FLAGS) -MM -MP -MT $(patsubst $(DEPSDIR)%.d, $(OBJSDIR)%.ll, $@)   $< > $@

# PROCESSDIR
objs/%.ll: $(PROCESSDIR)%.c
	$(CLANG) $(CLANG_FLAGS) -c $< -o $@

objs/bpf_multi_kprobe_v53.ll objs/bpf_multi_retkprobe_v53.ll:
	$(CLANG) $(CLANG_FLAGS) -D__LARGE_BPF_PROG -D__MULTI_KPROBE -c $< -o $@

objs/%_v53.ll:
	$(CLANG) $(CLANG_FLAGS) -D__LARGE_BPF_PROG -c $< -o $@

$(DEPSDIR)%.d: $(PROCESSDIR)%.c
	$(CLANG) $(CLANG_FLAGS) -MM -MP -MT $(patsubst $(DEPSDIR)%.d, $(OBJSDIR)%.ll, $@)   $< > $@

$(DEPSDIR)%_v53.d:
	$(CLANG) $(CLANG_FLAGS) -D__LARGE_BPF_PROG -MM -MP -MT $(patsubst $(DEPSDIR)%.d, $(OBJSDIR)%.ll, $@)   $< > $@

objs/bpf_multi_kprobe_v61.ll objs/bpf_multi_retkprobe_v61.ll:
	$(CLANG) $(CLANG_FLAGS) -D__LARGE_BPF_PROG -D__V61_BPF_PROG -D__MULTI_KPROBE -c $< -o $@

objs/%_v61.ll:
	$(CLANG) $(CLANG_FLAGS) -D__LARGE_BPF_PROG -D__V61_BPF_PROG -c $< -o $@

$(DEPSDIR)%_v61.d:
	$(CLANG) $(CLANG_FLAGS) -D__LARGE_BPF_PROG  -D__V61_BPF_PROG -MM -MP -MT $(patsubst $(DEPSDIR)%.d, $(OBJSDIR)%.ll, $@)   $< > $@

# BPFTESTDIR
objs/%.ll: $(BPFTESTDIR)%.c
	$(CLANG) $(CLANG_FLAGS) -c $< -o $@

$(DEPSDIR)%.d: $(BPFTESTDIR)%.c
	$(CLANG) $(CLANG_FLAGS) -MM -MP -MT $(patsubst $(DEPSDIR)%.d, $(OBJSDIR)%.ll, $@)   $< > $@

# CGROUPDIR
objs/%.ll: $(CGROUPDIR)%.c
	$(CLANG) $(CLANG_FLAGS) -c $< -o $@

$(DEPSDIR)%.d: $(CGROUPDIR)%.c
	$(CLANG) $(CLANG_FLAGS) -MM -MP -MT $(patsubst $(DEPSDIR)%.d, $(OBJSDIR)%.ll, $@)   $< > $@

# Remaining objects are built without mcpu=v2
objs/%.o: objs/%.ll
	$(LLC) $(LLC_FLAGS) -filetype=obj $< -o $@

# include dependencies
ifneq ($(MAKECMDGOALS),clean)
-include $(DEPS)
endif

clean:
	@$(ECHO_CLEAN)
	$(QUIET) $(foreach TARGET,$(SUBDIRS), \
		$(MAKE) -C $(TARGET) clean)
	$(QUIET)rm -f $(OBJSDIR)*.{o,ll,i,s}
	$(QUIET)rm -f $(DEPSDIR)*.d
