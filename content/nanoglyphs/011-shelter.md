+++
image_alt = "Wind chimes at home"
image_url = "/assets/images/nanoglyphs/011-shelter/chimes@2x.jpg"
published_at = 2020-03-27T18:02:51Z
title = "Sheltering, Twin Peaks; Asynchronous I/O, in Circles"
+++

Well, the world's changed since the last time we spoke. How many times in life can a person legitimately say that?

The windows are dark in San Francisco's museums, bars, and bookstores -- everywhere minus a few exceptions like grocery stores and restaurants with a take out business. Paradoxically, despite an ambient, pervasive feeling of malaise, the city feels nicer than than ever (we're still allowed out for exercise) -- traffic is lighter and more calm, the streets are quieter, the air is fresher [1].

Before the city's shelter in place order went out, Stripe's offices formally closed to all personnel. Before the office closure order went out, we'd been encouraged to work from home. That makes this my third week working out of my apartment, where before present circumstances I'd never spent even a fraction of this much contiguous time.

---

You're reading _Nanoglyph_, an experimental newsletter on software, ... and occasionally back yards and mountains. If you want to get straight to hard technology, scroll down a little ways for some close-to-the-metal content on the history of I/O APIs in the kernel. If you're pretty sure you never signed up for this wordy monstrosity and want to self-isolate from it _permanently_, unsubscribe in [one sterile click](%unsubscribe_url%).

If you're reading from the web, you can [subscribe here](https://nanoglyph-signup.brandur.org/). Allergic to the word "newsletter"? I don't blame you. I am too. Maybe you can think of it more like an async blog in the post-Google Reader age of consolidated social media. `ablog`? `bloguring`? We'll come back to that one.

---

A few years ago I moved to the Twin Peaks neighborhood of San Francisco. Just up the hill from the Castro and Cole Valley, and right across from the giant three-pronged TV/radio tower on top of Mount Sutro (Sutro Tower), still the most conspicuous landmark on the city's skyline, even with the notorious addition of Salesforce Tower. Even if you've never been to city, you might recognize its shape in the abstract in the designs of a thousand SF-centric T-shirts and coffee mugs.

![Sutro Tower](/assets/images/nanoglyphs/011-shelter/sutro-tower@2x.jpg)

Working from home has its ups and downs. My ergonomic situation is borderline dire -- my working posture isn't good in the best of times, and taking desk equipment out of the equation hasn't done it any favors. I'm changing sitting positions every 30 minutes (low desk, cross-legged on floor, couch, bed, cycle back to start) so my later life doesn't see me taking up residence in a chiropractor's office for live-in treatment. But a little hardship is balanced by an amazing overall reduction in travel-related stress. My commute is better than most of North America's, but even so, it's a tight ball of pain and anxiety that eats anywhere between 1 to 2 hours a day.

If you are going to be spending time at home, Twin Peaks isn't a bad place to do it. Even with surprisingly high apartment density and everyone holed up in their units, it's impressively quiet. I've been doing daily runs through our local trail system, across to Sutro's Open Space Reserve, and up to the top of the peaks themselves. My building's back yard (pictured at the top) makes a good place to write. An occasional meditative lap around the filled-in pool on the upper terrace helps to focus (pictured below; the compound's weirdest feature).

![Pool on the roof](/assets/images/nanoglyphs/011-shelter/pool@2x.jpg)

---

Despite rarely seeing another human being in the flesh, I manage to have 15 conversations about the virus before dinner. News is breaking all the time, and yet, changes slowly. Twitter is at its most unhelpful ever, which is saying something.

A thought that hits me hourly is that even in the face extraordinary events, this might just be one of the best self-development opportunities of all time. You know all those things that every adult claims to want to do but which is impossible because there's no time? Learning a language. Nailing down a healthy diet. Picking up an instrument. Writing a book. Learning to draw. This is it -- an extended snow day for most of humanity and the perfect time and excuse to stay home and do something constructive.

In theory anyway. So far I've been watching _The Sopranos_ and playing video games. I'm working on that.

---

## I/O classic (#io-classic)

Let's talk about asynchronous I/O in the Linux kernel. All the well-known disk syscalls in Linux like `read()`, `write()`, or `fsync()` are blocking -- the invoking program is paused while they do their work. They're all quite fast, made faster by the incredible high throughput disk technology we have these days, and made _even faster_ once the OS page cache is warm, so for most programs it doesn't matter. Callers also have the option of optimizing with [`posix_fadvise()`](http://man7.org/linux/man-pages/man2/posix_fadvise.2.html), which suggests to the kernel the sort of file data access they're going to engage in, and (probably [2]) have it warm up that page cache in advance.

There are however, classes of programs that see big benefits in performance by moving beyond synchronous I/O -- think something like a high throughput database, or web proxy that caches to disk. I find it easiest to think about with something like Node's event reactor, which is massively asynchronous, but running user code in only one place at any given time. If it were based naively on traditional file I/O functions, then any function calling `read()` would block everything else in the reactor until the operation completed. I/O-heavy Node apps would block constantly and have minimal practical concurrency.

But as it turns out, you can go far with synchronous I/O operations combined with some intelligent mitigations. In fact, Node _does_ call synchronous `read()`, but it does so through [`libuv`](http://docs.libuv.org/en/v1.x/design.html), which keeps a thread pool at the ready to parallelize these operations. By farming I/O out to a thread, Node makes [`fs.readFile()`](https://nodejs.org/api/fs.html#fs_fs_readfile_path_options_callback) asynchronous and non-blocking.

Another example of software that gets along fine on synchronous I/O is Postgres. It stays impressively fast by making liberal use of `posix_fadvise()` to warm the OS page cache, and having backends doing work in parallel across multiple OS processes, but it's all I/O classic under the hood.

---

## Early steps, and missteps (#early-steps)

POSIX has included an [`aio`](http://man7.org/linux/man-pages/man7/aio.7.html) (asynchronous I/O) API for some time that comes with async equivalents to file I/O functions: `aio_read()`, `aio_write()`, `aio_async()`, etc. However, because it operates in user space and uses threads to run async operations, it's not considered "truly" scalable by kernel standards, even if it is pretty fast.

The original `aio` implementation was followed by a kernel-based state machine encapsulated in the `io_*` class of functions. Clients would build a group of I/O operations to be processed asynchronously, send them en masse to [`io_submit()`](http://man7.org/linux/man-pages/man2/io_submit.2.html`) as a batch, then wait on results with [`io_getevents()`](http://man7.org/linux/man-pages/man2/io_getevents.2.html).

Like `aio_*`, `io_submit()` provides operations for file reading, writing, and fsync. More interestingly, it also provides the `IOCB_CMD_POLL` operation, which can be used to poll for ready sockets as an alternative to the more traditional select/poll/epoll APIs used by Nginx, `libuv`, and many other systems that need to manage asynchronous access across many waiting sockets. This [excellent article on `io_submit()`](https://blog.cloudflare.com/io_submit-the-epoll-alternative-youve-never-heard-about/) from CloudFlare makes a strong argument that `io_submit()` is preferable to epoll because its API is vastly simpler to use -- just push an array of relevant sockets into `io_submit()` then use `io_getevents()` to wait for completions. Here's a simple demonstration program from their post:

``` c
struct iocb cb = {.aio_fildes = sd,
                  .aio_lio_opcode = IOCB_CMD_POLL,
                  .aio_buf = POLLIN};
struct iocb *list_of_iocb[1] = {&cb};

r = io_submit(ctx, 1, list_of_iocb);
r = io_getevents(ctx, 1, 1, events, NULL);
```

That sounds pretty good so far, but `io_submit()` and company aren't without warts. Most notably, `io_submit()` is a still a blocking operation for most file I/O! (Remember, `io_submit()` is supposed to dispatch async operations whose results are waited on with `io_getevents()`, but not be synchronous itself.) It's possible to have `io_submit()` run truly asynchronously, but to do so it must be passed files that were opened with `O_DIRECT`.

`O_DIRECT` bypasses the operating system's page cache and other niceties, shifting onus onto the caller to ensure performance and correctness. It's an extremely low-level mechanism aimed at complex programs that need perfect control over what they're doing. Famously, [Linus hates it](https://lkml.org/lkml/2007/1/10/233), and the chances are that its legitimate uses in the real world are few and far between, which all puts a dramatic damper on the utility of `io_submit()`. This is all very poorly documented.

---

## The elegant symmetry of rings (#rings)

That brings us to today. A project that's been making amazing headway over the last few years and now included directly in the Linux kernel is `io_uring` ([this PDF](https://kernel.dk/io_uring.pdf) is its best self-contained description). It's a new system for asynchronous I/O that addresses the deficiencies of past generations and then some.

At its core are two ring buffers, the submission queue (SQ) and the completion queue (CQ). The fact that they're implemented as ring buffers (as opposed to any other type of queue) isn't all that important to an `io_uring` user, but is a nod to one of the project's guiding principles: efficiency. Recall that ring buffers allow each element in the queue to be used over and over again without allocating new memory. Each buffer tracks a head and tail, represented as monotonically increasing 32-bit integers. A simple modulo maps them to indexes in the buffers regardless of allocated size:

``` c
struct io_uring_sqe *sqe;
unsigned tail = sqring->tail;
unsigned index = tail & (*sqring->ring_mask);

/* put some new work into this submission
 * queue entry */
sqe = &sqring->sqes[index];
```

Client programs add work to the submission queue (SQ) by modifying entries at tail indexes, then updating that tail. Control is passed to the kernel, which consumes new entries and updates the buffer's head. Clients only ever update the queue's tail, and the kernel only ever updates the head, leading to minimal contention and little necessity for locking.

The roles are reversed for the completion queue (CQ). The kernel updates the queue's tail when a submission completes. Client programs read entries out of the queue and update the head as they finish consuming. Entries aren't guaranteed to be completed in the same order they were submitted, so each submission contains a `user_data` field/pointer to be specified, and which is made available in completed CQ entries to identify each piece of work:

``` c
struct io_uring_cqe {
   __u64 user_data;
   __s32 res;
   __u32 flags;
};
```

Although entries are unordered by default, `io_uring` makes some work easier by allowing programs to specify interdependencies. Setting the `IOSQE_IO_LINK` flag on an entry tells the kernel not to start the next entry before the current one is finished -- useful for issuing a write followed by an fsync for example. This little nicety is a nod to another one of `io_uring`'s design goals: ease of use -- the API should be intuitive and misuse should be difficult.

### Progressive enhancement with `liburing` (#liburing)

`io_uring`'s default API is designed to allow motivated users to squeeze every last drop of performance out of the new subsystem, but it's low level: just creating the initial state involves a manual call to `mmap()`, and the heads and tails of each queue are micromanaged at all times.

In an elegant example of progressive enhancement, `io_uring` also provides the `liburing` library, a simplified interface for basic use cases that cuts out almost every line of boilerplate. Here's a complete example of submitting an entry and waiting for it to finish:

``` c
struct io_uring ring;
io_uring_queue_init(ENTRIES, &ring, 0);

struct io_uring_sqe sqe;
struct io_uring_cqe cqe;

/* get an sqe and fill in a READV operation */
sqe = io_uring_get_sqe(&ring);
io_uring_prep_readv(sqe, fd, &iovec, 1, offset);

/* tell the kernel we have an sqe ready for
 * consumption */
io_uring_submit(&ring);

/* wait for the sqe to complete */
io_uring_wait_cqe(&ring, &cqe);

/* read and process cqe event */
app_handle_cqe(cqe);
io_uring_cqe_seen(&ring, cqe);
```

---

`io_uring` is brand new by the standards of syscall APIs, and of course Linux only, but it's showing huge promise in terms of usability, performance, and extensibility, all the while avoiding the pitfalls in which previous iterations have found themselves.

A significant next step will be to see which real-world programs find enough to like to adopt it. There may not be many yet, but they're coming. This [slide deck from Andres](https://anarazel.de/talks/2020-01-31-fosdem-aio/aio.pdf) talks about how baking a small `io_uring` prototype into Postgres yielded some very promising results, even when only minimally complete.

---

I'm overhauling my day-to-day by starting small. _Really_ small.

* **Cross-legged sitting:** I was taught to do this from the moment I entered preschool, but it never took and I've always regretted it. Eating meals on tatami in Japan, I was the only person in the room with a “cheat” pillow -- roughly equivalent in the realm of faux pas to asking for a fork. Ease back into some flexibility doing some cross-legged sitting every day (pictured below: my tea set seen from my new, lower perspective).

![Tea set -- made in Calgary and Japan](/assets/images/nanoglyphs/011-shelter/tea@2x.jpg)

* **Daily scheduling/routine:** With fewer commitments to be in certain places at certain times, my schedule's been on a collision course with its destiny as an unstructured, amorphous blob. It's not working. Go back to having a routine, even if not strictly as necessary.
* **Healthy meals:** I like carbs way too much. To do: eat food, not too much, mostly plants [3].
* **Technical reading:** Unlike fiction, technical reading requires time and concentration. And unlike fiction, it also makes you learn something. Do some every day.

If the small things go well, I'll work my way up to rehabilitating my derelict French later.

Take care.

[1] Air particulates in SF are reportedly down ~40% year over year.

[2] `posix_fadvise()`'s documentation is very careful to spell out that no requests made to it are binding. It sets an expectation on behalf of a program, which the kernel may or may not respect.

[3] See _Food Rules_ by Michael Pollan.
